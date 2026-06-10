import { InfoIcon, XIcon } from "lucide-react";
import React from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { toast } from "sonner";
import { AnchorReserveAlert } from "src/components/AnchorReserveAlert";
import AppHeader from "src/components/AppHeader";
import { CurrencyInputField } from "src/components/CurrencyInputField";
import { FeeRateField } from "src/components/FeeRateField";
import { InsufficientLightningBalanceAlert } from "src/components/InsufficientLightningBalanceAlert";
import Loading from "src/components/Loading";
import { MempoolAlert } from "src/components/MempoolAlert";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { Switch } from "src/components/ui/switch";
import { ONCHAIN_DUST_SATS } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useSwapInfo } from "src/hooks/useSwaps";
import {
  InitiateSwapRequest,
  RedeemOnchainFundsRequest,
  RedeemOnchainFundsResponse,
  SwapResponse,
} from "src/types";
import { request } from "src/utils/request";

export default function Onchain() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const [isSwap, setSwap] = React.useState(false);
  const address = state?.args?.address as string;
  const initialAmountSat = (state?.args?.amountSat as string | undefined) ?? "";
  const [amountSat, setAmountSat] = React.useState(initialAmountSat);

  React.useEffect(() => {
    if (!address) {
      navigate("/wallet/send");
    }
  }, [navigate, address]);

  if (!address) {
    return <Loading />;
  }

  return (
    <div className="grid gap-4">
      <AppHeader pageTitle="Send to On-chain" title="Send to On-chain" />
      <div className="grid gap-6 md:max-w-lg">
        <MempoolAlert />
        <div className="grid gap-2">
          <div className="text-sm font-medium">Recipient</div>
          <div className="flex items-center justify-between">
            <div className="flex flex-wrap gap-2 items-center font-mono text-sm">
              {address.match(/.{1,4}/g)?.map((word, index) => {
                if (index % 2 === 0) {
                  return (
                    <span key={index} className="text-foreground">
                      {word}
                    </span>
                  );
                } else {
                  return (
                    <span key={index} className="text-muted-foreground">
                      {word}
                    </span>
                  );
                }
              })}
            </div>
            <Link to="/wallet/send">
              <XIcon className="w-4 h-4 cursor-pointer text-muted-foreground" />
            </Link>
          </div>
        </div>
        {isSwap ? (
          <SwapForm
            address={address}
            setSwap={setSwap}
            amountSat={amountSat}
            setAmountSat={setAmountSat}
          />
        ) : (
          <OnchainForm
            address={address}
            setSwap={setSwap}
            amountSat={amountSat}
            setAmountSat={setAmountSat}
          />
        )}
      </div>
    </div>
  );
}

function OnchainForm({
  address,
  setSwap,
  amountSat,
  setAmountSat,
}: {
  address: string;
  amountSat: string;
  setAmountSat: React.Dispatch<React.SetStateAction<string>>;
  setSwap: React.Dispatch<React.SetStateAction<boolean>>;
}) {
  const navigate = useNavigate();
  const { data: balances } = useBalances();

  const [feeRate, setFeeRate] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (!balances) {
        return;
      }
      if (balances.onchain.spendableSat <= ONCHAIN_DUST_SATS) {
        throw new Error(
          "You currently don't have enough sats to pay for an on-chain transaction. Consider swapping from Lightning Balance."
        );
      }
      setLoading(true);
      const payload: RedeemOnchainFundsRequest = {
        toAddress: address,
        amountSat: +amountSat,
        feeRate: +feeRate,
      };
      const response = await request<RedeemOnchainFundsResponse>(
        "/api/wallet/redeem-onchain-funds",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(payload),
        }
      );
      if (!response?.txId) {
        throw new Error("No address in response");
      }
      navigate(`/wallet/send/onchain-success`, {
        state: {
          amountSat: +amountSat,
          txId: response.txId,
        },
        replace: true,
      });
      toast("Successfully broadcasted transaction");
    } catch (e) {
      toast.error("Failed to send payment", {
        description: "" + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  if (!balances) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-6">
      <CurrencyInputField
        id="amount"
        valueSat={amountSat}
        onValueSatChange={setAmountSat}
        minSat={ONCHAIN_DUST_SATS}
        maxSat={balances.onchain.spendableSat}
        required
        autoFocus
        contextRows={[
          {
            label: "On-chain available",
            amountSat: balances.onchain.spendableSat,
          },
        ]}
      />
      <div className="flex items-center justify-between">
        <Label htmlFor="swap" className="cursor-pointer">
          Swap from Lightning Balance
        </Label>
        <Switch id="swap" onCheckedChange={setSwap} />
      </div>
      <div className="grid gap-2 text-sm border-t pt-6">
        <FeeRateField feeRate={feeRate} onFeeRateChange={setFeeRate} />
      </div>
      {amountSat && +amountSat < 10_000 && (
        <Alert>
          <InfoIcon className="h-4 w-4" />
          <AlertTitle>Amount not ideal for On-chain transaction</AlertTitle>
          <AlertDescription>
            Small amounts can become unspendable when mempool fees increase.
            Consider using Lightning instead.
          </AlertDescription>
        </Alert>
      )}
      <AnchorReserveAlert amountSat={+amountSat} />
      <div className="flex gap-2">
        <LinkButton to="/wallet/send" variant="outline">
          Back
        </LinkButton>
        <LoadingButton loading={isLoading} type="submit" className="flex-1">
          Send
        </LoadingButton>
      </div>
    </form>
  );
}

function SwapForm({
  address,
  setSwap,
  amountSat,
  setAmountSat,
}: {
  address: string;
  amountSat: string;
  setAmountSat: React.Dispatch<React.SetStateAction<string>>;
  setSwap: React.Dispatch<React.SetStateAction<boolean>>;
}) {
  const navigate = useNavigate();
  const { data: balances } = useBalances();
  const { data: swapInfo } = useSwapInfo("out");

  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
      const payload: InitiateSwapRequest = {
        swapAmountSat: +amountSat,
        destination: address,
      };
      const swapOutResponse = await request<SwapResponse>("/api/swaps/out", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      if (!swapOutResponse) {
        throw new Error("Error swapping out");
      }
      navigate(`/wallet/swap/out/status/${swapOutResponse.swapId}`);
      toast("Initiated swap");
    } catch (e) {
      console.error(e);
      toast.error("Failed to send payment", {
        description: "" + e,
      });
    } finally {
      setLoading(false);
    }
  };
  const { data: recommendedFees } = useMempoolApi<{
    fastestFee: number;
  }>("/v1/fees/recommended");

  if (!balances || !swapInfo) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-6">
      <CurrencyInputField
        id="amount"
        valueSat={amountSat}
        onValueSatChange={setAmountSat}
        minSat={swapInfo.minAmountSat}
        maxSat={Math.min(
          swapInfo.maxAmountSat,
          balances.lightning.totalSpendableSat
        )}
        required
        autoFocus
        contextRows={[
          {
            label: "Lightning balance",
            amountSat: balances.lightning.totalSpendableSat,
          },
          {
            label: "Minimum",
            amountSat: swapInfo.minAmountSat,
          },
        ]}
      />
      <div className="flex items-center justify-between">
        <Label htmlFor="swap" className="cursor-pointer">
          Swap from Lightning Balance
        </Label>
        <Switch id="swap" checked onCheckedChange={setSwap} />
      </div>
      <div className="grid gap-2 text-sm border-t pt-6">
        <div className="flex items-center justify-between">
          <Label>On-chain Fee Rate</Label>
          <p>
            {recommendedFees?.fastestFee ? (
              <p>{recommendedFees?.fastestFee} sat/vB</p>
            ) : (
              <Loading className="w-4 h-4" />
            )}
          </p>
        </div>
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground">Swap Fee</p>
          <p>{swapInfo.albyServiceFee + swapInfo.boltzServiceFee}%</p>
        </div>
      </div>
      <InsufficientLightningBalanceAlert amountSat={+amountSat} />
      <div className="flex gap-2">
        <LinkButton to="/wallet/send" variant="outline">
          Back
        </LinkButton>
        <LoadingButton loading={isLoading} type="submit" className="flex-1">
          Send
        </LoadingButton>
      </div>
    </form>
  );
}
