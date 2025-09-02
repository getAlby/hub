import { ClipboardPasteIcon, MoveRightIcon, RefreshCwIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "src/components/ui/tabs";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useSwapInfo } from "src/hooks/useSwaps";
import { SwapResponse } from "src/types";
import { request } from "src/utils/request";

export default function Swap() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [tab, setTab] = useState(searchParams.get("type") || "in");

  useEffect(() => {
    const newTabValue = searchParams.get("type");
    if (newTabValue) {
      setTab(newTabValue);
      setSearchParams({});
    }
  }, [searchParams, setSearchParams]);

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Swap"
        contentRight={
          tab === "out" && (
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
      <Tabs value={tab} onValueChange={setTab} className="w-full max-w-lg">
        <TabsList className="w-full mb-4">
          <TabsTrigger value="in" className="flex gap-2 items-center w-full">
            Swap In
          </TabsTrigger>
          <TabsTrigger value="out" className="flex gap-2 items-center w-full">
            Swap Out
          </TabsTrigger>
        </TabsList>
        <TabsContent value="in">
          <SwapInForm />
        </TabsContent>
        <TabsContent value="out">
          <SwapOutForm />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function SwapInForm() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: swapInfo } = useSwapInfo("in");
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
      toast("Initiated swap");
    } catch (error) {
      toast.error("Failed to initiate swap", {
        description: (error as Error).message,
      });
    } finally {
      setLoading(false);
    }
  };

  if (!info || !balances || !swapInfo) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      <div>
        {hasChannelManagement &&
          parseInt(swapAmount || "0") * 1000 >=
            0.8 * balances.lightning.totalReceivable && (
            <LowReceivingCapacityAlert />
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
          min={swapInfo.minAmount}
          max={Math.min(
            swapInfo.maxAmount,
            (balances.lightning.totalReceivable / 1000) * 0.99
          )}
          onChange={(e) => setSwapAmount(e.target.value)}
          required
        />

        <div className="flex justify-between">
          {balances && (
            <div>
              <p className="text-xs text-muted-foreground">
                Receiving Capacity:{" "}
                {new Intl.NumberFormat().format(
                  Math.floor(balances.lightning.totalReceivable / 1000)
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
            Minimum: {new Intl.NumberFormat().format(swapInfo.minAmount)} sats
          </p>
        </div>
      </div>

      <div className="flex items-center justify-between border-t pt-4">
        <Label>Fee</Label>
        <p className="text-muted-foreground text-sm">
          {swapInfo.albyServiceFee + swapInfo.boltzServiceFee}% + on-chain fees
        </p>
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
  const { data: swapInfo } = useSwapInfo("out");
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
      toast("Initiated swap");
    } catch (error) {
      toast.error("Failed to initiate swap", {
        description: (error as Error).message,
      });
    } finally {
      setLoading(false);
    }
  };

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setDestination(text.trim());
  };

  if (!balances || !swapInfo) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      <div>
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
          min={swapInfo.minAmount}
          max={Math.min(
            swapInfo.maxAmount,
            Math.floor(balances.lightning.totalSpendable / 1000)
          )}
          onChange={(e) => setSwapAmount(e.target.value)}
          required
        />

        <div className="flex justify-between">
          {balances && (
            <p className="text-xs text-muted-foreground">
              Balance:{" "}
              {new Intl.NumberFormat().format(
                Math.floor(balances.lightning.totalSpendable / 1000)
              )}{" "}
              sats
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            Minimum: {new Intl.NumberFormat().format(swapInfo.minAmount)} sats
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
        <p className="text-muted-foreground text-sm">
          {swapInfo.albyServiceFee + swapInfo.boltzServiceFee}% + on-chain fees
        </p>
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
