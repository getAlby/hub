import {
  ArrowDownUpIcon,
  ClipboardPasteIcon,
  MoveRightIcon,
  XCircleIcon,
} from "lucide-react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useToast } from "src/components/ui/use-toast";
import { MIN_AUTO_SWAP_AMOUNT } from "src/constants";
import { useAutoSwapsConfig, useSwapFees } from "src/hooks/useSwaps";
import { request } from "src/utils/request";

export default function AutoSwap() {
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
      <div className="flex flex-col lg:flex-row gap-8 lg:gap-12 w-full">
        <div className="w-full lg:max-w-lg">
          <AutoSwapOutForm />
        </div>
        <ActiveSwaps />
      </div>
    </div>
  );
}

function AutoSwapOutForm() {
  const { toast } = useToast();
  const { mutate } = useAutoSwapsConfig();
  const { data: swapFees } = useSwapFees("out");
  const navigate = useNavigate();

  const [isInternalSwap, setInternalSwap] = useState(true);
  const [balanceThreshold, setBalanceThreshold] = useState("");
  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/wallet/autoswap/out", {
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
      navigate(`/wallet/swap/auto/success`, {
        state: {
          type: "out",
          isAutoSwap: true,
          balanceThreshold,
          amount: swapAmount,
        },
      });
      toast({
        title: "Saved successfully",
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
          min={MIN_AUTO_SWAP_AMOUNT}
          onChange={(e) => setSwapAmount(e.target.value)}
          required
        />
        <p className="text-xs text-muted-foreground">
          Minimum {new Intl.NumberFormat().format(MIN_AUTO_SWAP_AMOUNT)} sats
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
      <LoadingButton loading={loading}>Set Auto Swap Out</LoadingButton>
    </form>
  );
}

// TODO: Fix overflow in small screens
function ActiveSwaps() {
  const { toast } = useToast();
  const { data: swapConfig, mutate } = useAutoSwapsConfig();
  const { data: swapFees } = useSwapFees("out");

  const [loading, setLoading] = useState(false);

  const onDeactivate = async () => {
    try {
      setLoading(true);
      await request(`/api/wallet/autoswap/out`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      });
      toast({ title: "Deactivated successfully." });
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
    <div className="w-full flex flex-col gap-5 border-t-2 lg:border-t-0 pt-8 lg:pt-0">
      {swapConfig?.enabled && (
        <Card key={swapConfig.type}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="font-medium">Active Recurring Swap</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Alby Hub will try to perform a swap every time the balance reaches
              the threshold.
            </p>

            <div className="mt-6 space-y-4 text-sm">
              <div className="flex justify-between items-center">
                <span className="font-medium">Type</span>
                <span className="text-muted-foreground text-right">
                  {swapConfig.type == "out"
                    ? "Lightning to On-chain"
                    : "On-chain to Lightning"}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Destination</span>
                <span className="text-muted-foreground text-right">
                  {swapConfig.destination
                    ? swapConfig.destination
                    : swapConfig.type === "out"
                      ? "On-chain Balance"
                      : "Lightning Balance"}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Spending Balance Threshold</span>
                <span className="text-muted-foreground text-right">
                  {new Intl.NumberFormat().format(swapConfig.balanceThreshold)}{" "}
                  sats
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Swap amount</span>
                <span className="text-muted-foreground text-right">
                  {new Intl.NumberFormat().format(swapConfig.swapAmount)} sats
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Fee</span>
                {swapFees && (
                  <span className="text-muted-foreground text-right">
                    {swapFees.albyServiceFee + swapFees.boltzServiceFee}% +
                    on-chain fees
                  </span>
                )}
              </div>
            </div>
          </CardContent>
          <CardFooter className="flex justify-end">
            <Button
              onClick={() => onDeactivate()}
              disabled={loading}
              variant="outline"
            >
              <XCircleIcon className="h-4 w-4 mr-2" />
              Deactivate
            </Button>
          </CardFooter>
        </Card>
      )}
    </div>
  );
}
