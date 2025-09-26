import {
  AlertTriangleIcon,
  ExternalLinkIcon,
  InfoIcon,
  PencilIcon,
  XIcon,
} from "lucide-react";
import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import AnchorReserveAlert from "src/components/AnchorReserveAlert";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { MempoolAlert } from "src/components/MempoolAlert";
import { SpendingAlert } from "src/components/SpendingAlert";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { Switch } from "src/components/ui/switch";
import { ONCHAIN_DUST_SATS } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useSwapInfo } from "src/hooks/useSwaps";
import { RedeemOnchainFundsResponse, SwapResponse } from "src/types";
import { request } from "src/utils/request";

export default function Onchain() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const [isSwap, setSwap] = React.useState(true);
  const [amount, setAmount] = React.useState("");

  const address = state?.args?.address as string;

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
      <AppHeader title="Send to On-chain" />
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
            amount={amount}
            setAmount={setAmount}
          />
        ) : (
          <OnchainForm
            address={address}
            setSwap={setSwap}
            amount={amount}
            setAmount={setAmount}
          />
        )}
      </div>
    </div>
  );
}

function OnchainForm({
  address,
  setSwap,
  amount,
  setAmount,
}: {
  address: string;
  amount: string;
  setAmount: React.Dispatch<React.SetStateAction<string>>;
  setSwap: React.Dispatch<React.SetStateAction<boolean>>;
}) {
  const navigate = useNavigate();
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
  const { data: recommendedFees, error: mempoolError } = useMempoolApi<{
    fastestFee: number;
    halfHourFee: number;
    economyFee: number;
    minimumFee: number;
  }>("/v1/fees/recommended");

  const [feeRate, setFeeRate] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const [editFee, setEditFee] = React.useState(false);

  React.useEffect(() => {
    if (recommendedFees?.fastestFee) {
      setFeeRate(recommendedFees.fastestFee.toString());
    }
  }, [recommendedFees]);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (!balances) {
        return;
      }
      if (balances.onchain.spendable <= ONCHAIN_DUST_SATS) {
        throw new Error(
          "You currently don't have enough sats to pay for an on-chain transaction. Consider swapping from Spending Balance."
        );
      }
      setLoading(true);
      const response = await request<RedeemOnchainFundsResponse>(
        "/api/wallet/redeem-onchain-funds",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            toAddress: address,
            amount: +amount,
            feeRate: +feeRate,
          }),
        }
      );
      if (!response?.txId) {
        throw new Error("No address in response");
      }
      navigate(`/wallet/send/onchain-success`, {
        state: {
          amount: +amount,
          txId: response.txId,
        },
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

  if (!info || !balances || (!recommendedFees && !mempoolError)) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-6">
      <div className="grid gap-2">
        <Label htmlFor="amount">Amount</Label>
        <InputWithAdornment
          id="amount"
          type="number"
          value={amount}
          placeholder="Amount in Satoshi..."
          onChange={(e) => {
            setAmount(e.target.value.trim());
          }}
          min={ONCHAIN_DUST_SATS}
          max={balances.onchain.spendable}
          required
          autoFocus
          endAdornment={
            <FormattedFiatAmount amount={Number(amount)} className="mr-2" />
          }
        />
        <div className="flex justify-between text-muted-foreground text-xs sensitive slashed-zero">
          <div>
            On-chain Balance:{" "}
            {new Intl.NumberFormat().format(balances.onchain.spendable)} sats
          </div>
          <FormattedFiatAmount
            className="text-xs"
            amount={balances.onchain.spendable}
          />
        </div>
      </div>
      <div className="flex items-center justify-between">
        <Label htmlFor="swap" className="font-medium text-sm cursor-pointer">
          Swap from Spending Balance
        </Label>
        <Switch id="swap" onCheckedChange={setSwap} />
      </div>
      <div className="grid gap-2 text-sm border-t pt-6">
        {!editFee ? (
          <div className="flex items-center justify-between">
            <p className="text-muted-foreground">On-chain Fee Rate</p>
            <div
              className="flex items-center gap-2 cursor-pointer"
              onClick={() => setEditFee(true)}
            >
              {feeRate ? (
                <p>{feeRate} sat/vB</p>
              ) : (
                <Loading className="w-4 h-4" />
              )}
              <PencilIcon className="w-4 h-4" />
            </div>
          </div>
        ) : (
          <div className="grid gap-2">
            <Label htmlFor="fee-rate">Fee Rate (Sat/vB)</Label>
            {mempoolError && (
              <div className="text-muted-foreground text-xs flex gap-1 items-center">
                <AlertTriangleIcon className="h-3 w-3" />
                Failed to fetch fee estimates. Try refreshing the page.
              </div>
            )}
            <Input
              id="fee-rate"
              type="number"
              value={feeRate}
              step={1}
              required
              min={recommendedFees?.minimumFee || 1}
              onChange={(e) => {
                setFeeRate(e.target.value);
              }}
            />
            {recommendedFees && (
              <div className="flex items-center mt-2 gap-4">
                <Button
                  variant="positive"
                  className="rounded-full"
                  type="button"
                  onClick={() =>
                    setFeeRate(recommendedFees.economyFee.toString())
                  }
                >
                  Low priority: {recommendedFees.economyFee}
                </Button>{" "}
                <Button
                  variant="positive"
                  className="rounded-full"
                  type="button"
                  onClick={() =>
                    setFeeRate(recommendedFees.fastestFee.toString())
                  }
                >
                  High priority: {recommendedFees.fastestFee}
                </Button>{" "}
                <ExternalLink
                  to={info?.mempoolUrl}
                  className="text-muted-foreground underline flex items-center gap-2"
                >
                  View on Mempool
                  <ExternalLinkIcon className="w-4 h-4" />
                </ExternalLink>
              </div>
            )}
          </div>
        )}
      </div>
      {amount && +amount < 10_000 && (
        <Alert>
          <InfoIcon className="h-4 w-4" />
          <AlertTitle>Amount not ideal for On-chain transaction</AlertTitle>
          <AlertDescription>
            Small amounts can become unspendable when mempool fees increase.
            Consider using Lightning instead.
          </AlertDescription>
        </Alert>
      )}
      <AnchorReserveAlert amount={+amount} context="spend" />
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
  amount,
  setAmount,
}: {
  address: string;
  amount: string;
  setAmount: React.Dispatch<React.SetStateAction<string>>;
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
      const swapOutResponse = await request<SwapResponse>("/api/swaps/out", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: +amount,
          destination: address,
        }),
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
      <div className="grid gap-2">
        <Label htmlFor="amount">Amount</Label>
        <InputWithAdornment
          id="amount"
          type="number"
          value={amount}
          placeholder="Amount in Satoshi..."
          onChange={(e) => {
            setAmount(e.target.value.trim());
          }}
          min={swapInfo.minAmount}
          max={Math.min(
            swapInfo.maxAmount,
            Math.floor(balances.lightning.totalSpendable / 1000)
          )}
          required
          autoFocus
          endAdornment={
            <FormattedFiatAmount amount={Number(amount)} className="mr-2" />
          }
        />
        <div className="grid gap-1">
          <div className="flex justify-between text-xs text-muted-foreground sensitive slashed-zero">
            <div>
              Spending Balance:{" "}
              {new Intl.NumberFormat().format(
                Math.floor(balances.lightning.totalSpendable / 1000)
              )}{" "}
              sats
            </div>
            <FormattedFiatAmount
              className="text-xs"
              amount={Math.floor(balances.lightning.totalSpendable / 1000)}
            />
          </div>
          <div className="flex justify-between text-muted-foreground text-xs sensitive slashed-zero">
            <div>
              Minimum: {new Intl.NumberFormat().format(swapInfo.minAmount)} sats
            </div>
            <FormattedFiatAmount
              className="text-xs"
              amount={swapInfo.minAmount}
            />
          </div>
        </div>
      </div>
      <div className="flex items-center justify-between">
        <Label htmlFor="swap" className="font-medium text-sm cursor-pointer">
          Swap from Spending Balance
        </Label>
        <Switch id="swap" checked onCheckedChange={setSwap} />
      </div>
      <div className="grid gap-2 text-sm border-t pt-6">
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground">On-chain Fee Rate</p>
          <p>
            {/* ~{new Intl.NumberFormat().format(swapInfo.boltzNetworkFee)} sats */}
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
      <SpendingAlert amount={+amount} />
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
