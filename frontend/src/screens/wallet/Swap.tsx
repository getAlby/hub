import {
  ChevronDownIcon,
  ChevronUpIcon,
  ClipboardPasteIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useToast } from "src/components/ui/use-toast";
import { MIN_AUTO_SWAP_AMOUNT } from "src/constants";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { useSwaps } from "src/hooks/useSwaps";
import { cn } from "src/lib/utils";
import { request } from "src/utils/request";

export default function Swap() {
  const { data: swapsSettings } = useSwaps();
  const [swapType, setSwapType] = useState("out");

  useEffect(() => {
    const queryParams = new URLSearchParams(location.search);
    const swapType = queryParams.get("type");
    if (swapType) {
      setSwapType(swapType);
    }
  }, []);

  if (!swapsSettings) {
    return <Loading />;
  }

  return (
    <div className="w-full max-w-lg">
      <div className="flex items-center text-center text-foreground font-medium rounded-lg bg-muted p-1 mb-4">
        <div
          className={cn(
            "cursor-pointer rounded-md flex-1 py-1.5 text-sm",
            swapType == "in" && "bg-white font-bold"
          )}
          onClick={() => setSwapType("in")}
        >
          Swap In
        </div>
        <div
          className={cn(
            "cursor-pointer rounded-md flex-1 py-1.5 text-sm",
            swapType == "out" && "bg-white font-bold"
          )}
          onClick={() => setSwapType("out")}
        >
          Swap Out
        </div>
      </div>
      {swapType == "in" ? <SwapInForm /> : <SwapOutForm />}
    </div>
  );
}

function SwapInForm() {
  const { toast } = useToast();
  const { data: swapsSettings, mutate } = useSwaps();

  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState("");
  const [isInternalSwap, setInternalSwap] = useState(true);
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/wallet/swaps", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          // TODO: review conflict between swap in/out
          swapAmount: parseInt(swapAmount),
          destination: destination, // TODO: assume empty destination as dest -> "Spending Balance"
        }),
      });
      toast({ title: "Saved successfully." });
      await mutate();
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

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setDestination(text.trim());
  };

  if (!swapsSettings) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      <h2 className="text-muted-foreground">
        Swap on-chain funds into your lightning spending balance.
      </h2>
      <div className="grid gap-1.5">
        <Label>On-chain to Lightning swap</Label>
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
      <Label>Swap To</Label>
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
          <RadioGroupItem value="internal" id="internal" className="shrink-0" />
          <Label
            htmlFor="internal"
            className="text-primary font-medium cursor-pointer"
          >
            Spending balance
          </Label>
        </div>
        <div className="flex items-start space-x-2">
          <RadioGroupItem value="external" id="external" className="shrink-0" />
          <Label
            htmlFor="external"
            className="text-primary font-medium cursor-pointer"
          >
            External lightning wallet
          </Label>
        </div>
      </RadioGroup>
      {!isInternalSwap && (
        <div className="grid gap-1.5">
          <Label>Receiving lightning address</Label>
          <div className="flex gap-2 mb-4">
            <Input
              placeholder="satoshi@getalby.com"
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
      {/* TODO: Review fee for swap ins */}
      <div className="flex items-center justify-between border-t py-4">
        <Label>Fee</Label>
        <p className="text-muted-foreground text-sm">
          {swapsSettings.albyServiceFee + swapsSettings.boltzServiceFee}% +
          on-chain fees
        </p>
      </div>
      <LoadingButton loading={loading}>Swap In</LoadingButton>
    </form>
  );
}

function SwapOutForm() {
  const { toast } = useToast();
  // TODO: Optimize by setting this from the backend
  const { data: onchainAddress } = useOnchainAddress();
  const { data: swapsSettings, mutate } = useSwaps();
  const navigate = useNavigate();

  const [swapTo, setSwapTo] = useState("internal");
  const [balanceThreshold, setBalanceThreshold] = useState("");
  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState("");
  const [loading, setLoading] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [isRecurringSwap, setRecurringSwap] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      let txId;
      if (isRecurringSwap) {
        await request("/api/wallet/swaps", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            swapAmount: parseInt(swapAmount),
            destination: swapTo === "internal" ? onchainAddress : destination,
            balanceThreshold: parseInt(balanceThreshold),
          }),
        });
      } else {
        txId = await request<string>("/api/wallet/swap-out", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            swapAmount: parseInt(swapAmount),
            destination: swapTo === "internal" ? onchainAddress : destination,
          }),
        });
        if (!txId) {
          throw new Error("Error swapping out");
        }
      }
      navigate(`/wallet/swap/success`, {
        state: {
          type: "out",
          isRecurringSwap,
          txId,
          balanceThreshold,
          amount: swapAmount,
        },
      });
      toast({
        title: isRecurringSwap ? "Saved successfully" : "Initiated swap",
      });
      await mutate();
    } catch (error) {
      toast({
        title: isRecurringSwap
          ? "Failed to save auto swap settings"
          : "Failed to initiate swap",
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

  if (!onchainAddress || !swapsSettings) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      <h2 className="text-muted-foreground">
        Swap lightning funds into your on-chain balance.
      </h2>
      <div className="grid gap-1.5">
        <Label>Lightning to On-chain swap</Label>
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
      <Label>Swap To</Label>
      <RadioGroup
        defaultValue="normal"
        value={swapTo}
        onValueChange={(val) => {
          setSwapTo(val);
          if (val == "internal") {
            setDestination(onchainAddress);
          } else {
            setDestination("");
          }
        }}
        className="flex gap-4 flex-row"
      >
        <div className="flex items-start space-x-2 mb-2">
          <RadioGroupItem value="internal" id="internal" className="shrink-0" />
          <Label
            htmlFor="internal"
            className="text-primary font-medium cursor-pointer"
          >
            Alby Hub on-chain balance
          </Label>
        </div>
        <div className="flex items-start space-x-2">
          <RadioGroupItem value="external" id="external" className="shrink-0" />
          <Label
            htmlFor="external"
            className="text-primary font-medium cursor-pointer"
          >
            External on-chain wallet
          </Label>
        </div>
      </RadioGroup>
      {swapTo == "external" && (
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

      <Button
        type="button"
        variant="link"
        className="text-muted-foreground text-xs"
        onClick={() => setShowAdvanced((current) => !current)}
      >
        {showAdvanced ? (
          <ChevronUpIcon className="w-4 h-4 mr-2" />
        ) : (
          <ChevronDownIcon className="w-4 h-4 mr-2" />
        )}
        Advanced Options
      </Button>

      {showAdvanced && (
        <>
          <div className="flex items-top space-x-2">
            <Checkbox
              id="public-channel"
              checked={isRecurringSwap}
              onCheckedChange={() => setRecurringSwap(!isRecurringSwap)}
              className="mr-2"
            />
            <div className="grid gap-1.5 leading-none">
              <Label
                htmlFor="public-channel"
                className="flex items-center gap-2"
              >
                Set it as recurring swap
              </Label>
              <p className="text-xs text-muted-foreground">
                Enable if you want Alby Hub to automatically swap funds every
                time a set threshold is reached.
              </p>
            </div>
          </div>
          {isRecurringSwap && (
            <div className="mt-2 grid gap-1.5">
              <Label>Swap threshold</Label>
              <Input
                type="number"
                placeholder="Amount in satoshis"
                value={balanceThreshold}
                onChange={(e) => setBalanceThreshold(e.target.value)}
                required
              />
              <p className="text-xs text-muted-foreground">
                Alby Hub will try to perform a swap to on-chain everytime
                lightning balance reaches this threshold.
              </p>
            </div>
          )}
        </>
      )}

      <div className="flex items-center justify-between border-t py-4">
        <Label>Fee</Label>
        <p className="text-muted-foreground text-sm">
          {swapsSettings.albyServiceFee + swapsSettings.boltzServiceFee}% +
          on-chain fees
        </p>
      </div>
      <LoadingButton loading={loading}>
        {isRecurringSwap ? "Enable Auto Swap-outs" : "Swap Out"}
      </LoadingButton>
    </form>
  );
}
