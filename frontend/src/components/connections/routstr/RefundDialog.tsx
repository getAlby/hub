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
import { getRoutstrdBalance, refundFromHub } from "src/hooks/useRoutstrd";
import { request } from "src/utils/request";
import { handleRequestError } from "src/utils/handleRequestError";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRefundComplete: () => void;
  /** Routstr connection app ID — sats go back to this app's isolated wallet */
  appId: number;
};

const MIN_REFUND_SATS = 10;
/**
 * Input-fee buffer: the Cashu mint charges an input fee per proof
 * (input_fee_ppk, typically 100msat/proof → ~1 sat for small wallets)
 * plus output/swap fees (~1 sat). Without this buffer the melt can
 * fail with "Not enough funds available to send" even when the balance
 * covers invoice + fee_reserve.
 */
const INPUT_FEE_BUFFER = 2;

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
  const [mintFee, setMintFee] = useState<number | null>(null);
  const [loadingBalances, setLoadingBalances] = useState(false);
  const [refundPhase, setRefundPhase] = useState("");

  // Cache last-known values so subsequent opens feel instant
  const lastBalance = useRef<number | null>(null);
  const lastFee = useRef<number | null>(null);
  const hasEverLoaded = useRef(false);

  const doLoadBalances = useCallback(async () => {
    // 1. Get Cashu balance from daemon
    const balResult = await getRoutstrdBalance();
    const mints = balResult?.balances ? Object.keys(balResult.balances) : [];
    const mintUrl = balResult?.activeMint || mints[0];
    const walletBal = balResult?.balances
      ? Object.values(balResult.balances).reduce((a, b) => a + b, 0)
      : 0;

    // 2. Get real fee: create a Hub invoice, query the mint for melt quote on it
    let fee = 0;
    if (mintUrl && walletBal > 0) {
      try {
        const invResult = await request<{ invoice: string }>("/api/invoices", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          // App-scoped: even the fee-quote invoice belongs to the Routstr
          // wallet, so the main wallet is never involved (it's unpaid —
          // used only to query the mint's fee_reserve).
          body: JSON.stringify({ amount: walletBal * 1000, appId }),
        });
        const invoice = invResult?.invoice;
        if (invoice) {
          const mtResp = await fetch(
            `${mintUrl.replace(/\/+$/, "")}/v1/melt/quote/bolt11`,
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
      } catch {
        fee = 0;
      }
    }

    // Update state and cache
    setWalletBalance(walletBal);
    setMintFee(fee);
    lastBalance.current = walletBal;
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
  const minRequired = Math.max(
    MIN_REFUND_SATS,
    effectiveFee + INPUT_FEE_BUFFER + 1
  );
  const sendAmount =
    walletBalance !== null && walletBalance >= minRequired
      ? walletBalance - effectiveFee - INPUT_FEE_BUFFER
      : 0;
  const canRefund = walletBalance !== null && walletBalance >= minRequired;

  const handleRefund = async () => {
    if (!canRefund || sendAmount <= 0) {
      return;
    }
    setIsProcessing(true);
    setStep("processing");
    try {
      setRefundPhase(`Refunding ${sendAmount} sats...`);
      const refunded = await refundFromHub(sendAmount, appId);
      toast.success(`Refunded ${refunded} sats to your Routstr wallet`);
      setStep("done");
      onRefundComplete();
    } catch (error) {
      handleRequestError("Refund failed", error);
      setStep("confirm");
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
                      {walletBalance ?? "?"} sats
                    </span>
                  </div>
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
                        Minimum refund is {MIN_REFUND_SATS} sats.{" "}
                        {walletBalance !== null &&
                        walletBalance < MIN_REFUND_SATS
                          ? `Top up to at least ${MIN_REFUND_SATS} sats first.`
                          : `Network fee (~${effectiveFee} sats) + input fees (~${INPUT_FEE_BUFFER} sats) leave nothing to refund.`}
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
