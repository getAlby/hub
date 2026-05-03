import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { FixedFloatSwapInFlow } from "src/components/FixedFloatSwapInFlow";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import { MempoolAlert } from "src/components/MempoolAlert";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useSwapInfo } from "src/hooks/useSwaps";
import {
  CreateInvoiceRequest,
  InitiateSwapRequest,
  SwapResponse,
  Transaction,
} from "src/types";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request";

export default function ReceiveOnchain() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: swapInfo } = useSwapInfo("in");
  const { data: recommendedFees, error: mempoolError } = useMempoolApi<{
    fastestFee: number;
    halfHourFee: number;
    economyFee: number;
    minimumFee: number;
  }>("/v1/fees/recommended");
  const navigate = useNavigate();

  const [swapFrom, setSwapFrom] = useState<"bitcoin" | "crypto">("bitcoin");
  const [swapAmountSat, setSwapAmountSat] = useState("");
  const [loading, setLoading] = useState(false);
  const [feeRate, setFeeRate] = useState("");
  const [cryptoTransaction, setCryptoTransaction] =
    useState<Transaction | null>(null);

  useEffect(() => {
    if (recommendedFees?.fastestFee) {
      setFeeRate(recommendedFees.fastestFee.toString());
    }
  }, [recommendedFees]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      if (swapFrom === "crypto") {
        const tx = await request<Transaction>("/api/invoices", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            amountMsat: (parseInt(swapAmountSat) || 0) * 1000,
            description: "Fixed Float swap",
          } as CreateInvoiceRequest),
        });
        if (!tx?.invoice) {
          throw new Error("Failed to create invoice");
        }
        setCryptoTransaction(tx);
        openLink(
          `https://ff.io/?to=BTCLN&address=${encodeURIComponent(tx.invoice)}&ref=qnnjvywb`
        );
        toast("Initiated swap");
        return;
      }

      const payload: InitiateSwapRequest = {
        swapAmountSat: parseInt(swapAmountSat),
      };
      const swapInResponse = await request<SwapResponse>("/api/swaps/in", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      if (!swapInResponse) {
        throw new Error("Error swapping in");
      }
      navigate(`/wallet/swap/in/status/${swapInResponse.swapId}`);
      toast("Initiated swap");
    } catch (error) {
      toast.error("Failed to initiate swap", {
        description: (error as Error).message,
      });
    } finally {
      setLoading(false);
    }
  };

  if (!info || !balances || !swapInfo || (!recommendedFees && !mempoolError)) {
    return <Loading />;
  }

  const isCryptoReceiveState =
    swapFrom === "crypto" && cryptoTransaction !== null;

  return (
    <div className="grid gap-5">
      <AppHeader
        pageTitle="Receive from On-chain"
        title="Receive from On-chain"
      />
      <div className="w-full max-w-lg grid gap-6">
        <MempoolAlert />
        <form onSubmit={onSubmit} className="flex flex-col gap-6">
          {!isCryptoReceiveState && (
            <>
              {hasChannelManagement &&
                parseInt(swapAmountSat || "0") * 1000 >=
                  0.8 * balances.lightning.totalReceivableMsat && (
                  <LowReceivingCapacityAlert />
                )}
              <div className="grid gap-1.5">
                <Label>Amount</Label>
                <InputWithAdornment
                  type="number"
                  autoFocus
                  placeholder="Amount in satoshis"
                  value={swapAmountSat}
                  min={
                    swapFrom === "bitcoin" ? swapInfo.minAmountSat : undefined
                  }
                  max={
                    swapFrom === "bitcoin"
                      ? Math.min(
                          swapInfo.maxAmountSat,
                          balances.lightning.totalReceivableSat * 0.99
                        )
                      : balances.lightning.totalReceivableSat * 0.99
                  }
                  onChange={(e) => setSwapAmountSat(e.target.value)}
                  required
                  endAdornment={
                    <FormattedFiatAmount
                      amountSat={+swapAmountSat}
                      className="mr-2"
                    />
                  }
                />
                <div className="grid">
                  <div className="flex justify-between text-muted-foreground text-xs sensitive slashed-zero">
                    <div>
                      Receiving Capacity:{" "}
                      <FormattedBitcoinAmount
                        amountMsat={balances.lightning.totalReceivableMsat}
                      />
                    </div>
                    <FormattedFiatAmount
                      className="text-xs"
                      amountSat={balances.lightning.totalReceivableSat}
                    />
                  </div>
                </div>
              </div>
              <div className="flex flex-col gap-4">
                <Label>Swap from</Label>
                <RadioGroup
                  defaultValue="bitcoin"
                  value={swapFrom}
                  onValueChange={(value) => {
                    setSwapFrom(value as "bitcoin" | "crypto");
                  }}
                  className="flex gap-4 flex-row"
                >
                  <div className="flex items-start space-x-2 mb-2">
                    <RadioGroupItem
                      value="bitcoin"
                      id="bitcoin"
                      className="shrink-0"
                    />
                    <Label htmlFor="bitcoin" className="cursor-pointer">
                      Bitcoin
                    </Label>
                  </div>
                  <div className="flex items-start space-x-2">
                    <RadioGroupItem
                      value="crypto"
                      id="crypto"
                      className="shrink-0"
                    />
                    <Label htmlFor="crypto" className="cursor-pointer">
                      Other Cryptocurrency
                    </Label>
                  </div>
                </RadioGroup>
              </div>
            </>
          )}

          {swapFrom === "bitcoin" ? (
            <BitcoinSwapFlow
              feeRate={feeRate}
              loading={loading}
              swapFee={swapInfo.albyServiceFee + swapInfo.boltzServiceFee}
            />
          ) : (
            <FixedFloatSwapInFlow
              loading={loading}
              transaction={cryptoTransaction}
              resetLabel="Receive Another Payment"
              onReset={() => {
                setCryptoTransaction(null);
                setSwapAmountSat("");
              }}
            />
          )}
        </form>
      </div>
    </div>
  );
}

function BitcoinSwapFlow({
  feeRate,
  loading,
  swapFee,
}: {
  feeRate: string;
  loading: boolean;
  swapFee: number;
}) {
  return (
    <>
      <div className="border-t pt-4 text-sm grid gap-2">
        <div className="flex items-center justify-between">
          <Label>On-chain Fee</Label>
          {feeRate ? (
            <p className="text-muted-foreground">{feeRate} sat/vB</p>
          ) : (
            <Loading className="w-4 h-4" />
          )}
        </div>
        <div className="flex items-center justify-between">
          <Label>Swap Fee</Label>
          <p className="text-muted-foreground">{swapFee}%</p>
        </div>
      </div>

      <LoadingButton className="w-full" loading={loading}>
        Continue
      </LoadingButton>
    </>
  );
}
