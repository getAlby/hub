import {
  AlertTriangleIcon,
  ClipboardPasteIcon,
  MoveRightIcon,
  RefreshCwIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useToast } from "src/components/ui/use-toast";
import { MIN_AUTO_SWAP_AMOUNT } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useSwapFees } from "src/hooks/useSwaps";
import { cn } from "src/lib/utils";
import { SwapResponse } from "src/types";
import { request } from "src/utils/request";

export default function Swap() {
  const [swapType, setSwapType] = useState("in");

  useEffect(() => {
    const queryParams = new URLSearchParams(location.search);
    const swapType = queryParams.get("type");
    if (swapType) {
      setSwapType(swapType);
    }
  }, []);

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Swap"
        contentRight={
          swapType === "out" && (
            <Link to="/wallet/swap/auto">
              <ResponsiveButton
                variant="outline"
                icon={RefreshCwIcon}
                text="Auto Swap"
              />
            </Link>
          )
        }
      />
      <div className="w-full max-w-lg">
        <div className="flex items-center text-center text-foreground font-medium rounded-lg bg-muted p-1">
          <div
            className={cn(
              "cursor-pointer rounded-md flex-1 py-1.5 text-sm",
              swapType == "in" && "text-foreground bg-background font-semibold"
            )}
            onClick={() => setSwapType("in")}
          >
            Swap In
          </div>
          <div
            className={cn(
              "cursor-pointer rounded-md flex-1 py-1.5 text-sm",
              swapType == "out" && "text-foreground bg-background font-semibold"
            )}
            onClick={() => setSwapType("out")}
          >
            Swap Out
          </div>
        </div>
        {swapType == "in" ? <SwapInForm /> : <SwapOutForm />}
      </div>
    </div>
  );
}

function SwapInForm() {
  const { toast } = useToast();
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: swapFees } = useSwapFees("in");
  const navigate = useNavigate();

  const [swapAmount, setSwapAmount] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      const swapInResponse = await request<SwapResponse>("/api/swaps/in", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: parseInt(swapAmount),
        }),
      });
      if (!swapInResponse) {
        throw new Error("Error swapping in");
      }
      navigate(`/wallet/swap/in/status/${swapInResponse.swapId}`);
      toast({ title: "Initiated swap" });
    } catch (error) {
      toast({
        title: "Failed to initiate swap",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  if (!info || !balances) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      <div className="mt-6">
        {hasChannelManagement &&
          parseInt(swapAmount || "0") * 1000 >=
          0.8 * balances.lightning.totalReceivable && (
            <Alert className="mb-6">
              <AlertTriangleIcon className="h-4 w-4" />
              <AlertTitle>Low receiving capacity</AlertTitle>
              <AlertDescription>
                You likely won't be able to receive payments until you{" "}
                <Link className="underline" to="/channels/incoming">
                  increase your receiving capacity.
                </Link>
              </AlertDescription>
            </Alert>
          )}
        <h2 className="font-medium text-foreground flex items-center gap-1">
          On-chain <MoveRightIcon /> Lightning
        </h2>
        <p className="mt-1 text-muted-foreground">
          Swap on-chain funds into your lightning spending balance.
        </p>
      </div>
      <div className="grid gap-1.5">
        <Label>Swap amount</Label>
        <Input
          type="number"
          autoFocus
          placeholder="Amount in satoshis"
          value={swapAmount}
          min={MIN_AUTO_SWAP_AMOUNT}
          max={(balances.lightning.totalReceivable / 1000) * 0.99}
          onChange={(e) => setSwapAmount(e.target.value)}
          required
        />

        <div className="flex justify-between">
          {balances && (
            <div>
              <p className="text-xs text-muted-foreground">
                Receiving Capacity:{" "}
                {new Intl.NumberFormat().format(
                  balances.lightning.totalReceivable / 1000
                )}{" "}
                sats{" "}
                <Link className="underline" to="/channels/incoming">
                  increase
                </Link>
              </p>
              <p className="text-xs text-muted-foreground">
                On-Chain Balance:{" "}
                {new Intl.NumberFormat().format(balances.onchain.spendable)}{" "}
                sats
              </p>
            </div>
          )}
          <p className="text-xs text-muted-foreground">
            Minimum: {new Intl.NumberFormat().format(MIN_AUTO_SWAP_AMOUNT)} sats
          </p>
        </div>
      </div>

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
      <div className="grid gap-2">
        <LoadingButton className="w-full" loading={loading}>
          Swap In
        </LoadingButton>
        <p className="text-xs text-muted-foreground text-center">
          powered by{" "}
          <span className="font-medium text-foreground">boltz.exchange</span>
        </p>
      </div>
    </form>
  );
}

function SwapOutForm() {
  const { toast } = useToast();
  const { data: swapFees } = useSwapFees("out");
  const navigate = useNavigate();
  const { data: balances } = useBalances();

  const [isInternalSwap, setInternalSwap] = useState(true);
  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      const swapOutResponse = await request<SwapResponse>("/api/swaps/out", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: parseInt(swapAmount),
          destination,
        }),
      });
      if (!swapOutResponse) {
        throw new Error("Error swapping out");
      }
      navigate(`/wallet/swap/out/status/${swapOutResponse.swapId}`);
      toast({ title: "Initiated swap" });
    } catch (error) {
      toast({
        title: "Failed to initiate swap",
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
      <div className="mt-6">
        <h2 className="font-medium text-foreground flex items-center gap-1">
          Lightning <MoveRightIcon /> On-chain
        </h2>
        <p className="mt-1 text-muted-foreground">
          Swap bitcoin lightning into your on-chain balance.
        </p>
      </div>
      <div className="grid gap-1.5">
        <Label>Swap amount</Label>
        <Input
          type="number"
          autoFocus
          placeholder="Amount in satoshis"
          value={swapAmount}
          min={MIN_AUTO_SWAP_AMOUNT}
          onChange={(e) => setSwapAmount(e.target.value)}
          required
        />

        <div className="flex justify-between">
          {balances && (
            <p className="text-xs text-muted-foreground">
              Balance:{" "}
              {new Intl.NumberFormat().format(
                balances.lightning.totalSpendable / 1000
              )}{" "}
              sats
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            Minimum: {new Intl.NumberFormat().format(MIN_AUTO_SWAP_AMOUNT)} sats
          </p>
        </div>
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
          <div className="flex gap-2">
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
      <div className="grid gap-2">
        <LoadingButton className="w-full" loading={loading}>
          Swap Out
        </LoadingButton>
        <p className="text-xs text-muted-foreground text-center">
          powered by{" "}
          <span className="font-medium text-foreground">boltz.exchange</span>
        </p>
      </div>
    </form>
  );
}
