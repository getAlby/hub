import {
  ClipboardPasteIcon,
  ExternalLinkIcon,
  InfoIcon,
  MoveRightIcon,
  RefreshCwIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import { FixedFloatSwapInFlow } from "src/components/FixedFloatSwapInFlow";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import ResponsiveLinkButton from "src/components/ResponsiveLinkButton";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { Separator } from "src/components/ui/separator";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "src/components/ui/tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useSwapInfo } from "src/hooks/useSwaps";
import {
  CreateInvoiceRequest,
  InitiateSwapRequest,
  SwapResponse,
  Transaction,
} from "src/types";
import { openLink } from "src/utils/openLink";
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
        pageTitle="Swap"
        title="Swap"
        contentRight={
          tab === "out" && (
            <ResponsiveLinkButton
              to="/wallet/swap/auto"
              variant="outline"
              icon={RefreshCwIcon}
              text="Auto Swap"
            />
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
  const [swapFrom, setSwapFrom] = useState<"internal" | "external" | "crypto">(
    "external"
  );
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: swapInfo } = useSwapInfo("in");
  const { data: channels } = useChannels();
  const navigate = useNavigate();

  const [swapAmountSat, setSwapAmountSat] = useState("");
  const [loading, setLoading] = useState(false);
  const [cryptoTransaction, setCryptoTransaction] =
    useState<Transaction | null>(null);

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
      navigate(
        `/wallet/swap/in/status/${swapInResponse.swapId}${swapFrom === "internal" ? "?internal=true" : ""}`
      );
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

  const spendableOnchainBalanceSatWithAnchorReserves = Math.max(
    balances.onchain.spendableSat - (channels?.length || 0) * 25000,
    0
  );
  const isInternalSwap = swapFrom === "internal";
  const isCryptoSwappingState =
    swapFrom === "crypto" && cryptoTransaction !== null;

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      {!isCryptoSwappingState && (
        <>
          <div>
            <h2 className="font-medium text-foreground flex items-center gap-1">
              On-chain <MoveRightIcon /> Lightning
            </h2>
            <p className="mt-1 text-muted-foreground">
              Swap on-chain funds into your lightning spending balance.
            </p>
          </div>
          <div className="grid gap-1.5">
            {hasChannelManagement &&
              parseInt(swapAmountSat || "0") * 1000 >=
                0.8 * balances.lightning.totalReceivableMsat && (
                <div className="mb-4">
                  <LowReceivingCapacityAlert />
                </div>
              )}

            <Label>Swap amount</Label>
            <Input
              type="number"
              autoFocus
              placeholder="Amount in satoshis"
              value={swapAmountSat}
              min={swapFrom !== "crypto" ? swapInfo.minAmountSat : undefined}
              max={
                swapFrom === "crypto"
                  ? balances.lightning.totalReceivableSat * 0.99
                  : Math.min(
                      swapInfo.maxAmountSat,
                      ...(isInternalSwap
                        ? [spendableOnchainBalanceSatWithAnchorReserves]
                        : []),
                      balances.lightning.totalReceivableSat * 0.99
                    )
              }
              onChange={(e) => setSwapAmountSat(e.target.value)}
              required
            />

            <div className="flex justify-between">
              {balances && (
                <div>
                  <p className="text-xs text-muted-foreground">
                    Receiving Capacity:{" "}
                    <FormattedBitcoinAmount
                      amountMsat={balances.lightning.totalReceivableMsat}
                    />
                  </p>
                  {isInternalSwap && (
                    <p className="text-xs text-muted-foreground flex items-center justify-center gap-1">
                      Spendable On-Chain Balance:{" "}
                      <FormattedBitcoinAmount
                        amountMsat={
                          spendableOnchainBalanceSatWithAnchorReserves * 1000
                        }
                      />
                      {!!channels?.length && (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <div className="flex flex-row gap-1 items-center text-muted-foreground">
                                <InfoIcon className="h-3 w-3 shrink-0" />
                              </div>
                            </TooltipTrigger>
                            <TooltipContent>
                              To ensure you can close channels, you need to set
                              aside at least{" "}
                              <FormattedBitcoinAmount
                                amountMsat={channels.length * 25000 * 1000}
                              />{" "}
                              on-chain. Your total on-chain balance is{" "}
                              <FormattedBitcoinAmount
                                amountMsat={
                                  balances.onchain.spendableSat * 1000
                                }
                              />
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}
                    </p>
                  )}
                </div>
              )}
            </div>
          </div>
          <div className="flex flex-col gap-4">
            <Label>Swap from</Label>
            <RadioGroup
              defaultValue="external"
              value={swapFrom}
              onValueChange={(value: "internal" | "external" | "crypto") => {
                setSwapFrom(value);
              }}
              className="flex gap-4 flex-wrap"
            >
              <div className="flex items-start space-x-2 mb-2">
                <RadioGroupItem
                  value="internal"
                  id="internal"
                  className="shrink-0"
                />
                <Label htmlFor="internal" className="cursor-pointer">
                  On-chain balance
                </Label>
              </div>
              <div className="flex items-start space-x-2">
                <RadioGroupItem
                  value="external"
                  id="external"
                  className="shrink-0"
                />
                <Label htmlFor="external" className="cursor-pointer">
                  External on-chain wallet
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

      {swapFrom === "crypto" ? (
        <FixedFloatSwapInFlow
          loading={loading}
          transaction={cryptoTransaction}
          resetLabel="Swap Another Amount"
          onReset={() => {
            setCryptoTransaction(null);
            setSwapAmountSat("");
          }}
        />
      ) : (
        <>
          <div className="flex items-center justify-between border-t pt-4">
            <Label>Fee</Label>
            <p className="text-muted-foreground text-sm">
              {swapInfo.albyServiceFee + swapInfo.boltzServiceFee}% + on-chain
              fees
            </p>
          </div>
          <div className="grid gap-2">
            <LoadingButton className="w-full" loading={loading}>
              Swap In
            </LoadingButton>
            <p className="text-xs text-muted-foreground text-center">
              powered by{" "}
              <span className="font-medium text-foreground">
                boltz.exchange
              </span>
            </p>
          </div>
        </>
      )}
    </form>
  );
}

function SwapOutForm() {
  const { data: swapInfo } = useSwapInfo("out");
  const navigate = useNavigate();
  const { data: balances } = useBalances();

  const [isInternalSwap, setInternalSwap] = useState(true);
  const [swapAmountSat, setSwapAmountSat] = useState("");
  const [destination, setDestination] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      const payload: InitiateSwapRequest = {
        swapAmountSat: parseInt(swapAmountSat),
        destination,
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
          value={swapAmountSat}
          min={swapInfo.minAmountSat}
          max={Math.min(
            swapInfo.maxAmountSat,
            balances.lightning.totalSpendableSat
          )}
          onChange={(e) => setSwapAmountSat(e.target.value)}
          required
        />

        <div className="flex justify-between">
          {balances && (
            <p className="text-xs text-muted-foreground">
              Balance:{" "}
              <FormattedBitcoinAmount
                amountMsat={balances.lightning.totalSpendableMsat}
              />
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            Minimum:{" "}
            <FormattedBitcoinAmount amountMsat={swapInfo.minAmountSat * 1000} />
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
            <Label htmlFor="internal" className="cursor-pointer">
              On-chain balance
            </Label>
          </div>
          <div className="flex items-start space-x-2">
            <RadioGroupItem
              value="external"
              id="external"
              className="shrink-0"
            />
            <Label htmlFor="external" className="cursor-pointer">
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
      <Separator className="my-2" />
      <FixedFloatButton from="BTCLN" className="w-full" variant="secondary">
        <ExternalLinkIcon className="size-4" />
        Swap out to other Cryptocurrency
      </FixedFloatButton>
    </form>
  );
}
