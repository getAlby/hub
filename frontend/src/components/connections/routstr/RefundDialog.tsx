import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import {
  getRoutstrdBalance,
  getRoutstrdKeyBalances,
  PartialRefundError,
  reclaimProviderTokens,
  refundFromHub,
} from "src/hooks/useRoutstrd";
import { request } from "src/utils/request";
import { handleRequestError } from "src/utils/handleRequestError";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRefundComplete: () => void;
  /** Routstr connection app ID — sats go back to this app's isolated wallet */
  appId: number;
};

export function RefundDialog({
  open,
  onOpenChange,
  onRefundComplete,
  appId,
}: Props) {
  const [isProcessing, setIsProcessing] = useState(false);
  const [step, setStep] = useState<"confirm" | "processing" | "done">(
    "confirm"
  );
  const [walletBalance, setWalletBalance] = useState<number | null>(null);
  // Prepaid provider tokens (apikey:* entries) — not in the wallet, but
  // reclaimable via the daemon /refund before melting.
  const [providerTokens, setProviderTokens] = useState(0);
  const [mintUrl, setMintUrl] = useState<string | null>(null);
  const [mintFee, setMintFee] = useState<number | null>(null);
  const [loadingBalances, setLoadingBalances] = useState(false);
  const [refundPhase, setRefundPhase] = useState("");

  // Cache last-known values so subsequent opens feel instant
  const lastBalance = useRef<number | null>(null);
  const lastProviderTokens = useRef(0);
  const lastFee = useRef<number | null>(null);
  // Balance at the last mint fee quote — avoids re-quoting (and creating a
  // lingering app-scoped invoice) on every dialog open / silent refresh.
  const lastQuotedBalance = useRef(0);
  const hasEverLoaded = useRef(false);

  const doLoadBalances = useCallback(async () => {
    // 1. Get Cashu balance from daemon
    const balResult = await getRoutstrdBalance();
    const mints = balResult?.balances ? Object.keys(balResult.balances) : [];
    const activeMint = balResult?.activeMint || mints[0] || null;
    const walletBal = balResult?.balances
      ? Object.values(balResult.balances).reduce((a, b) => a + b, 0)
      : 0;

    // 2. Prepaid provider tokens (apikey:* entries in /keys/balance). These
    // are NOT in the wallet and cannot be melted directly — the refund
    // reclaims them first (daemon /refund), so show them as refundable.
    let providerFloat: number;
    try {
      const keyResult = await getRoutstrdKeyBalances();
      providerFloat = (keyResult?.keys || [])
        .filter((k) => k.id !== "wallet")
        .reduce((sum, k) => sum + (k.balance || 0), 0);
    } catch {
      providerFloat = 0;
    }
    const totalRefundable = walletBal + providerFloat;

    // 3. Get real fee: create a Hub invoice, query the mint for melt quote on it.
    // Only re-quote when the balance moved materially — every quote creates an
    // app-scoped invoice that lingers as a pending transaction, and the dialog
    // refreshes on every open (open + silent background refresh).
    let fee = 0;
    if (
      lastQuotedBalance.current !== 0 &&
      lastFee.current !== null &&
      Math.abs(totalRefundable - lastQuotedBalance.current) < 10
    ) {
      fee = lastFee.current;
    } else if (activeMint && totalRefundable > 0) {
      try {
        const invResult = await request<{ invoice: string }>("/api/invoices", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          // App-scoped: even the fee-quote invoice belongs to the Routstr
          // wallet, so the main wallet is never involved (it's unpaid —
          // used only to query the mint's fee_reserve).
          body: JSON.stringify({ amount: totalRefundable * 1000, appId }),
        });
        const invoice = invResult?.invoice;
        if (invoice) {
          const mtResp = await fetch(
            `${activeMint.replace(/\/+$/, "")}/v1/melt/quote/bolt11`,
            {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ request: invoice, unit: "sat" }),
            }
          );
          if (mtResp.ok) {
            const quote = await mtResp.json();
            fee = typeof quote.fee_reserve === "number" ? quote.fee_reserve : 0;
          }
        }
        lastQuotedBalance.current = totalRefundable;
      } catch {
        fee = 0;
      }
    }

    // Update state and cache
    setWalletBalance(walletBal);
    setProviderTokens(providerFloat);
    setMintUrl(activeMint);
    setMintFee(fee);
    lastBalance.current = walletBal;
    lastProviderTokens.current = providerFloat;
    lastFee.current = fee;
    hasEverLoaded.current = true;
  }, [appId]);

  const loadBalances = useCallback(async () => {
    setLoadingBalances(true);
    try {
      await doLoadBalances();
    } catch {
      // Non-critical
    } finally {
      setLoadingBalances(false);
    }
  }, [doLoadBalances]);

  const loadBalancesSilent = useCallback(async () => {
    try {
      await doLoadBalances();
    } catch {
      // Silently fail — cached values are already displayed
    }
  }, [doLoadBalances]);

  useEffect(() => {
    if (open) {
      setStep("confirm");
      setIsProcessing(false);

      if (hasEverLoaded.current) {
        // Show cache immediately, refresh silently in background
        setWalletBalance(lastBalance.current);
        setProviderTokens(lastProviderTokens.current);
        setMintFee(lastFee.current);
        setLoadingBalances(false);
        loadBalancesSilent();
      } else {
        // First load — show spinner
        loadBalances();
      }
    }
  }, [open, loadBalances, loadBalancesSilent]);

  const effectiveFee = mintFee ?? 0;
  // Melt anything that leaves at least 1 sat after the fee, so the wallet
  // can be drained to zero (no arbitrary 10-sat floor).
  const minRequired = Math.max(1, effectiveFee + 1);
  // Refundable = wallet + reclaimable provider tokens (reclaimed before melt).
  // The refund drains the whole wallet: net = balance − fee (no extra buffer
  // is held back, so nothing is left behind as dust).
  const totalRefundable =
    walletBalance !== null ? walletBalance + providerTokens : null;
  const sendAmount =
    totalRefundable !== null && totalRefundable >= minRequired
      ? Math.floor(totalRefundable - effectiveFee)
      : 0;
  const canRefund = totalRefundable !== null && totalRefundable >= minRequired;

  const handleRefund = async () => {
    if (!canRefund || sendAmount <= 0) {
      return;
    }
    setIsProcessing(true);
    setStep("processing");
    try {
      if (providerTokens > 0 && mintUrl) {
        setRefundPhase("Reclaiming provider tokens...");
        try {
          await reclaimProviderTokens(mintUrl);
        } catch {
          // Provider refund failed — continue with the wallet balance only
        }
      }
      if (!mintUrl) {
        throw new Error("No active mint to melt from");
      }
      setRefundPhase("Refunding wallet balance...");
      // refundFromHub drains the ENTIRE Cashu wallet (fee quoted fresh each
      // pass), so the wallet ends at zero — no dust left behind.
      const refunded = await refundFromHub(appId, mintUrl);
      toast.success(`Refunded ${refunded} sats to your Routstr wallet`);
      setStep("done");
      onRefundComplete();
    } catch (error) {
      if (error instanceof PartialRefundError && error.totalRefunded > 0) {
        // Earlier drain passes already moved sats — report the partial
        // outcome instead of a blanket failure.
        toast.warning(
          `Partially refunded ${error.totalRefunded} sats to your Routstr wallet (${error.message})`
        );
        setStep("done");
        onRefundComplete();
      } else {
        handleRequestError("Refund failed", error);
        setStep("confirm");
      }
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        {step === "confirm" && (
          <>
            <DialogHeader>
              <DialogTitle>Refund API Balance</DialogTitle>
            </DialogHeader>

            <div className="rounded-lg border border-border/60 bg-muted/10 p-4 space-y-2 text-sm">
              {loadingBalances ? (
                <div className="flex items-center gap-2 text-muted-foreground">
                  <span className="h-3 w-3 rounded-full border border-border animate-spin border-t-transparent" />
                  Calculating fees…
                </div>
              ) : (
                <>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">
                      API Key balance
                    </span>
                    <span className="font-medium tabular-nums">
                      {totalRefundable ?? "?"} sats
                    </span>
                  </div>
                  {walletBalance !== null && providerTokens > 0 && (
                    <div className="flex justify-between text-[10px] text-muted-foreground/60">
                      <span>
                        {walletBalance} in wallet + {providerTokens.toFixed(2)}{" "}
                        in provider tokens
                      </span>
                      <span>reclaimed on refund</span>
                    </div>
                  )}
                  {walletBalance !== null && (
                    <div className="flex justify-between text-muted-foreground">
                      <span>Network fee (fee_reserve)</span>
                      <span className="tabular-nums">~{effectiveFee} sats</span>
                    </div>
                  )}
                  <hr className="border-border/40" />
                  {walletBalance !== null && (
                    <div className="flex justify-between font-medium">
                      <span>Net to Routstr wallet</span>
                      <span className="tabular-nums">
                        {canRefund ? sendAmount : 0} sats
                      </span>
                    </div>
                  )}
                  {walletBalance !== null && !canRefund && (
                    <div className="text-center space-y-1">
                      <p className="text-xs text-amber-600 dark:text-amber-400">
                        Balance too low
                      </p>
                      <p className="text-[10px] text-muted-foreground/60">
                        {totalRefundable === 0
                          ? `Network fee (~${effectiveFee} sats) leaves nothing to refund.`
                          : `Top up to at least ${minRequired} sats first.`}
                      </p>
                    </div>
                  )}
                </>
              )}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <LoadingButton
                loading={isProcessing}
                onClick={handleRefund}
                disabled={!canRefund}
              >
                Confirm Refund
              </LoadingButton>
            </DialogFooter>
          </>
        )}

        {step === "processing" && (
          <>
            <DialogHeader>
              <DialogTitle>Refunding…</DialogTitle>
            </DialogHeader>
            <div className="py-8 text-center text-sm text-muted-foreground">
              {refundPhase}
            </div>
          </>
        )}

        {step === "done" && (
          <>
            <DialogHeader>
              <DialogTitle>Refund complete</DialogTitle>
            </DialogHeader>
            <DialogFooter>
              <Button onClick={() => onOpenChange(false)}>Close</Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
