import { ClipboardPasteIcon, MoveRightIcon, RefreshCwIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useToast } from "src/components/ui/use-toast";
import { MIN_AUTO_SWAP_AMOUNT } from "src/constants";
import { useSwaps } from "src/hooks/useSwaps";
import { cn } from "src/lib/utils";
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
          <Link to="/wallet/swap/auto">
            <ResponsiveButton
              variant="outline"
              icon={RefreshCwIcon}
              text="Auto Swap"
            />
          </Link>
        }
      />
      <div className="w-full max-w-lg">
        <div className="flex items-center text-center text-foreground font-medium rounded-lg bg-muted p-1">
          <div
            className={cn(
              "cursor-pointer rounded-md flex-1 py-1.5 text-sm",
              swapType == "in" && "bg-white font-semibold"
            )}
            onClick={() => setSwapType("in")}
          >
            Swap In
          </div>
          <div
            className={cn(
              "cursor-pointer rounded-md flex-1 py-1.5 text-sm",
              swapType == "out" && "bg-white font-semibold"
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
  const { data: swapSettings } = useSwaps();
  const swapInSettings = swapSettings?.find((s) => s.type === "in");

  const [swapAmount, setSwapAmount] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/wallet/swap/in", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: parseInt(swapAmount),
        }),
      });
      toast({ title: "Initiated swap" });
    } catch (error) {
      toast({
        title: "Saving swap settings failed",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      <div className="mt-6">
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

      {/* TODO: Review fee for swap ins */}
      <div className="flex items-center justify-between border-t pt-4">
        <Label>Fee</Label>
        {swapInSettings ? (
          <p className="text-muted-foreground text-sm">
            {swapInSettings.albyServiceFee + swapInSettings.boltzServiceFee}% +
            on-chain fees
          </p>
        ) : (
          <Loading />
        )}
      </div>
      <LoadingButton loading={loading}>Swap In</LoadingButton>
    </form>
  );
}

function SwapOutForm() {
  const { toast } = useToast();
  const { data: swapSettings } = useSwaps();
  const swapOutSettings = swapSettings?.find((s) => s.type === "out");
  const navigate = useNavigate();

  const [isInternalSwap, setInternalSwap] = useState(true);
  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      const txId = await request<string>("/api/wallet/swap/out", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: parseInt(swapAmount),
          destination,
        }),
      });
      if (!txId) {
        throw new Error("Error swapping out");
      }
      navigate(`/wallet/swap/success`, {
        state: {
          type: "out",
          txId,
          amount: swapAmount,
        },
      });
      toast({
        title: "Initiated swap",
      });
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
        {swapOutSettings ? (
          <p className="text-muted-foreground text-sm">
            {swapOutSettings.albyServiceFee + swapOutSettings.boltzServiceFee}%
            + on-chain fees
          </p>
        ) : (
          <Loading />
        )}
      </div>
      <LoadingButton loading={loading}>Swap Out</LoadingButton>
    </form>
  );
}
