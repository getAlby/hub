import { InfoIcon, XIcon } from "lucide-react";
import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LinkButton } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Switch } from "src/components/ui/switch";
import { useToast } from "src/components/ui/use-toast";
import { MIN_SWAP_AMOUNT, ONCHAIN_DUST_SATS } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { useSwapFees } from "src/hooks/useSwaps";
import { RedeemOnchainFundsResponse, SwapResponse } from "src/types";
import { request } from "src/utils/request";

export default function Onchain() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const [isSwap, setSwap] = React.useState(false);

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
        <div className="grid gap-2">
          <div className="text-sm font-medium">Recipient</div>
          <div className="flex items-center justify-between">
            <div className="flex flex-wrap gap-2 items-center font-mono">
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
              <XIcon className="w-4 h-4 cursor-pointer" />
            </Link>
          </div>
        </div>
        {isSwap ? (
          <SwapForm address={address} setSwap={setSwap} />
        ) : (
          <OnchainForm address={address} setSwap={setSwap} />
        )}
      </div>
    </div>
  );
}

function OnchainForm({
  address,
  setSwap,
}: {
  address: string;
  setSwap: React.Dispatch<React.SetStateAction<boolean>>;
}) {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { data: balances } = useBalances();
  const { data: recommendedFees, error: mempoolError } = useMempoolApi<{
    fastestFee: number;
    halfHourFee: number;
    economyFee: number;
    minimumFee: number;
  }>("/v1/fees/recommended");

  const [amount, setAmount] = React.useState("");
  const [feeRate, setFeeRate] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

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
      toast({
        title: "Successfully broadcasted transaction",
      });
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to send payment",
        description: "" + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  if (!balances || (!recommendedFees && !mempoolError)) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-6">
      <div className="grid gap-2">
        <Label htmlFor="amount">Amount</Label>
        <Input
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
          className="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
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
      <div className="grid gap-2 text-sm border-t pt-4">
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground">On-chain Fee</p>
          {feeRate ? <p>{feeRate} sat/vB</p> : <Loading className="w-4 h-4" />}
        </div>
      </div>
      {amount && +amount < 10_000 && (
        <Alert>
          <InfoIcon className="h-4 w-4" />
          <AlertTitle>Amount not ideal for On-chain transaction</AlertTitle>
          <AlertDescription>
            Small amounts can become unspendable during fee spikes. Consider
            using Lightning instead.
          </AlertDescription>
        </Alert>
      )}
      <div className="flex gap-2">
        <LinkButton to="/wallet/send" variant="outline">
          Back
        </LinkButton>
        <LoadingButton
          loading={isLoading}
          type="submit"
          className="w-full md:w-fit"
        >
          Send
        </LoadingButton>
      </div>
    </form>
  );
}

function SwapForm({
  address,
  setSwap,
}: {
  address: string;
  setSwap: React.Dispatch<React.SetStateAction<boolean>>;
}) {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { data: balances } = useBalances();
  const { data: swapFees } = useSwapFees("out");

  const [amount, setAmount] = React.useState("");
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
          isSending: true,
        }),
      });
      if (!swapOutResponse) {
        throw new Error("Error swapping out");
      }
      navigate(`/wallet/swap/out/status/${swapOutResponse.swapId}`);
      toast({ title: "Initiated swap" });
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to send payment",
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
      <div className="grid gap-2">
        <Label htmlFor="amount">Amount</Label>
        <Input
          id="amount"
          type="number"
          value={amount}
          placeholder="Amount in Satoshi..."
          onChange={(e) => {
            setAmount(e.target.value.trim());
          }}
          min={MIN_SWAP_AMOUNT}
          max={Math.floor(balances.lightning.totalSpendable / 1000)}
          required
          autoFocus
          className="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
          endAdornment={
            <FormattedFiatAmount amount={Number(amount)} className="mr-2" />
          }
        />
        <div className="grid gap-1">
          <div className="flex justify-between text-muted-foreground text-xs sensitive slashed-zero">
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
            <div>Minimum: 50000 sats</div>
            <FormattedFiatAmount className="text-xs" amount={50000} />
          </div>
        </div>
      </div>
      <div className="flex items-center justify-between">
        <Label htmlFor="swap" className="font-medium text-sm cursor-pointer">
          Swap from Spending Balance
        </Label>
        <Switch id="swap" checked onCheckedChange={setSwap} />
      </div>
      <div className="grid gap-2 text-sm border-t pt-4">
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground">On-chain Fee</p>
          {swapFees ? (
            <p>
              ~{new Intl.NumberFormat().format(swapFees.boltzNetworkFee)} sats
            </p>
          ) : (
            <Loading className="w-4 h-4" />
          )}
        </div>
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground">Swap Fee</p>
          {swapFees ? (
            <p>{swapFees.albyServiceFee + swapFees.boltzServiceFee}%</p>
          ) : (
            <Loading className="w-4 h-4" />
          )}
        </div>
      </div>
      <div className="flex gap-2">
        <LinkButton to="/wallet/send" variant="outline">
          Back
        </LinkButton>
        <LoadingButton
          loading={isLoading}
          type="submit"
          className="w-full md:w-fit"
        >
          Send
        </LoadingButton>
      </div>
    </form>
  );
}
