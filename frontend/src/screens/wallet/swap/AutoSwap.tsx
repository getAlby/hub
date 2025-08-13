import {
  ArrowDownUpIcon,
  ClipboardPasteIcon,
  ClockIcon,
  MoveRightIcon,
  XCircleIcon,
} from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useAutoSwapsConfig, useSwapFees } from "src/hooks/useSwaps";
import { AutoSwapConfig } from "src/types";
import { request } from "src/utils/request";

export default function AutoSwap() {
  const { data: swapConfig } = useAutoSwapsConfig();

  if (!swapConfig) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Auto Swap Out"
        contentRight={
          <Link to="/wallet/swap">
            <ResponsiveButton
              variant="outline"
              icon={ArrowDownUpIcon}
              text="Swap"
            />
          </Link>
        }
      />
      <div className="w-full lg:max-w-lg min-w-0">
        {swapConfig.enabled ? (
          <ActiveSwapOutConfig swapConfig={swapConfig} />
        ) : (
          <AutoSwapOutForm />
        )}
      </div>
    </div>
  );
}

function AutoSwapOutForm() {
  const { toast } = useToast();
  const { data: balances } = useBalances();
  const { mutate } = useAutoSwapsConfig();
  const { data: swapFees } = useSwapFees("out");

  const [isInternalSwap, setInternalSwap] = useState(true);
  const [balanceThreshold, setBalanceThreshold] = useState("");
  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/autoswap", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: parseInt(swapAmount),
          balanceThreshold: parseInt(balanceThreshold),
          destination,
        }),
      });
      toast({
        title: "Auto swap enabled successfully",
      });
      await mutate();
    } catch (error) {
      toast({
        title: "Failed to save auto swap settings",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setDestination(text.trim());
  };

  if (!balances || !swapFees) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      <div>
        <h2 className="font-medium text-foreground flex items-center gap-1">
          Lightning <MoveRightIcon /> On-chain
        </h2>
        <p className="mt-1 text-muted-foreground">
          Setup automatic swap of lightning funds into your on-chain balance
          every time a set threshold is reached.
        </p>
        <p className="mt-2 text-muted-foreground flex gap-2 items-center text-sm">
          <ClockIcon className="w-4 h-4" />
          Swaps will be made once per hour
        </p>
      </div>

      <div className="grid gap-1.5">
        <Label>Spending balance threshold</Label>
        <Input
          type="number"
          placeholder="Amount in satoshis"
          value={balanceThreshold}
          onChange={(e) => setBalanceThreshold(e.target.value)}
          required
        />
        <p className="text-xs text-muted-foreground">
          Swap out as soon as this amount is reached
        </p>
      </div>

      <div className="grid gap-1.5">
        <Label>Swap amount</Label>
        <Input
          type="number"
          placeholder="Amount in satoshis"
          value={swapAmount}
          min={swapFees.minAmount}
          max={Math.min(
            swapFees.maxAmount,
            Math.floor(balances.lightning.totalSpendable / 1000)
          )}
          onChange={(e) => setSwapAmount(e.target.value)}
          required
        />
        <p className="text-xs text-muted-foreground">
          Minimum {new Intl.NumberFormat().format(swapFees.minAmount)} sats
        </p>
      </div>
      <div className="flex flex-col gap-4">
        <Label>Swap to</Label>
        <RadioGroup
          defaultValue="normal"
          value={isInternalSwap ? "internal" : "external"}
          onValueChange={() => {
            setDestination("");
            setInternalSwap(!isInternalSwap);
          }}
          className="flex gap-4 flex-row"
        >
          <div className="flex items-start space-x-2 mb-2">
            <RadioGroupItem
              value="internal"
              id="internal"
              className="shrink-0"
            />
            <Label
              htmlFor="internal"
              className="text-primary font-medium cursor-pointer"
            >
              On-chain balance
            </Label>
          </div>
          <div className="flex items-start space-x-2">
            <RadioGroupItem
              value="external"
              id="external"
              className="shrink-0"
            />
            <Label
              htmlFor="external"
              className="text-primary font-medium cursor-pointer"
            >
              External on-chain wallet
            </Label>
          </div>
        </RadioGroup>
      </div>
      {!isInternalSwap && (
        <div className="grid gap-1.5">
          <Label>Receiving on-chain address</Label>
          <div className="flex gap-2 mb-4">
            <Input
              placeholder="bc1..."
              value={destination}
              onChange={(e) => setDestination(e.target.value)}
              required
            />
            <Button
              type="button"
              variant="outline"
              className="px-2"
              onClick={paste}
            >
              <ClipboardPasteIcon className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      <div className="flex items-center justify-between border-t pt-4">
        <Label>Fee</Label>
        {swapFees ? (
          <p className="text-muted-foreground text-sm">
            {swapFees.albyServiceFee + swapFees.boltzServiceFee}% + on-chain
            fees
          </p>
        ) : (
          <Loading />
        )}
      </div>
      <div className="grid gap-1">
        <LoadingButton className="w-full" loading={loading}>
          Begin Auto Swap
        </LoadingButton>
        <p className="text-xs text-muted-foreground text-right">
          powered by{" "}
          <ExternalLink
            to="https://boltz.exchange"
            className="font-medium text-foreground"
          >
            boltz.exchange
          </ExternalLink>
        </p>
      </div>
    </form>
  );
}

function ActiveSwapOutConfig({ swapConfig }: { swapConfig: AutoSwapConfig }) {
  const { toast } = useToast();
  const { mutate } = useAutoSwapsConfig();
  const { data: swapFees } = useSwapFees("out");

  const [loading, setLoading] = useState(false);

  const onDeactivate = async () => {
    try {
      setLoading(true);
      await request(`/api/autoswap`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      });
      toast({ title: "Deactivated auto swap successfully" });
      await mutate();
    } catch (error) {
      toast({
        title: "Deactivating auto swaps failed",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <h2 className="font-medium text-foreground flex items-center gap-1">
        Active Lightning <MoveRightIcon /> On-chain Swap
      </h2>
      <p className="mt-1 text-muted-foreground">
        Alby Hub will try to perform a swap every time the balance reaches the
        threshold.
      </p>
      <p className="mt-2 text-muted-foreground flex gap-2 items-center text-sm">
        <ClockIcon className="w-4 h-4" />
        Swaps will be made once per hour
      </p>

      <div className="my-6 space-y-4 text-sm">
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium">Type</span>
          <span className="truncate text-muted-foreground text-right">
            Lightning to On-chain
          </span>
        </div>
        <div className="flex justify-between items-center gap-2">
          <div className="font-medium">Destination</div>
          <div className="truncate text-muted-foreground text-right">
            {swapConfig.destination
              ? swapConfig.destination
              : "On-chain Balance"}
          </div>
        </div>
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium truncate">
            Spending Balance Threshold
          </span>
          <span className="shrink-0 text-muted-foreground text-right">
            {new Intl.NumberFormat().format(swapConfig.balanceThreshold)} sats
          </span>
        </div>
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium truncate">Swap amount</span>
          <span className="shrink-0 text-muted-foreground text-right">
            {new Intl.NumberFormat().format(swapConfig.swapAmount)} sats
          </span>
        </div>
        <div className="flex justify-between items-center gap-2">
          <span className="font-medium">Fee</span>
          {swapFees ? (
            <span className="truncate text-muted-foreground text-right">
              {swapFees.albyServiceFee + swapFees.boltzServiceFee}% + on-chain
              fees
            </span>
          ) : (
            <Loading className="w-4 h-4" />
          )}
        </div>
      </div>
      <Button
        onClick={() => onDeactivate()}
        disabled={loading}
        variant="outline"
      >
        <XCircleIcon className="h-4 w-4 mr-2" />
        Deactivate Auto Swap
      </Button>
    </>
  );
}
