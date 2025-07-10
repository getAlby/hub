import { ClipboardPasteIcon, XCircleIcon } from "lucide-react";
import { useState } from "react";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
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
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { useSwaps } from "src/hooks/useSwaps";
import { request } from "src/utils/request";

function Swaps() {
  const { toast } = useToast();
  // TODO: Optimize by setting this from the backend
  const { data: onchainAddress } = useOnchainAddress();
  const { data: swapsSettings, mutate } = useSwaps();

  const [swapTo, setSwapTo] = useState("hub");
  const [balanceThreshold, setBalanceThreshold] = useState("");
  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/settings/swaps", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount: parseInt(swapAmount),
          balanceThreshold: parseInt(balanceThreshold),
          destination: swapTo === "hub" ? onchainAddress : destination,
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

  const onDeactivate = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/settings/swaps", {
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

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setDestination(text.trim());
  };

  if (!onchainAddress || !swapsSettings) {
    return <Loading />;
  }

  return (
    <>
      <SettingsHeader
        title="Swaps"
        description="Automatically swap lightning to on-chain funds."
      />
      {!swapsSettings.enabled ? (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          <div className="grid gap-1.5">
            <Label>Spending balance threshold</Label>
            <Input
              type="number"
              placeholder="Swap out as soon as this amount is reached"
              value={balanceThreshold}
              onChange={(e) => setBalanceThreshold(e.target.value)}
            />
          </div>
          <div className="grid gap-1.5">
            <Label>Swap amount</Label>
            <Input
              type="number"
              placeholder="How much do you want to swap out?"
              value={swapAmount}
              min={MIN_AUTO_SWAP_AMOUNT}
              onChange={(e) => setSwapAmount(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Minimum {new Intl.NumberFormat().format(MIN_AUTO_SWAP_AMOUNT)}{" "}
              sats
            </p>
          </div>
          <Label>Destination</Label>
          <RadioGroup
            defaultValue="normal"
            value={swapTo}
            onValueChange={(val) => {
              setSwapTo(val);
              if (val == "hub") {
                setDestination(onchainAddress);
              } else {
                setDestination("");
              }
            }}
            className="flex gap-4 flex-row"
          >
            <div className="flex items-start space-x-2 mb-2">
              <RadioGroupItem value="hub" id="hub" className="shrink-0" />
              <Label
                htmlFor="hub"
                className="text-primary font-medium cursor-pointer"
              >
                Alby Hub on-chain balance
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
          {swapTo == "external" && (
            <div className="grid gap-1.5">
              <Label>Receiving on-chain address</Label>
              <div className="flex gap-2 mb-4">
                <Input
                  placeholder="bc1..."
                  value={destination}
                  onChange={(e) => setDestination(e.target.value)}
                />
                <Button
                  type="button"
                  variant="outline"
                  className="px-2"
                  onClick={paste}
                >
                  <ClipboardPasteIcon className="size-4" />
                </Button>
              </div>
            </div>
          )}

          <div className="flex items-center justify-between border-t py-4">
            <Label>Fee</Label>
            <p className="text-muted-foreground text-sm">
              {swapsSettings.albyServiceFee + swapsSettings.boltzServiceFee}% +
              on-chain fees
            </p>
          </div>
          <LoadingButton
            loading={loading}
            disabled={
              !balanceThreshold || (swapTo == "external" && !destination)
            }
          >
            Enable Auto Swaps
          </LoadingButton>
        </form>
      ) : (
        <Card className="w-full hidden md:block self-start">
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
                  Lightning to On-chain
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Destination</span>
                <span className="text-muted-foreground text-right">
                  {swapsSettings.destination}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Spending Balance Threshold</span>
                <span className="text-muted-foreground text-right">
                  {new Intl.NumberFormat().format(
                    swapsSettings.balanceThreshold
                  )}{" "}
                  sats
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Swap amount</span>
                <span className="text-muted-foreground text-right">
                  {new Intl.NumberFormat().format(swapsSettings.swapAmount)}{" "}
                  sats
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Fee</span>
                <span className="text-muted-foreground text-right">
                  {swapsSettings.albyServiceFee + swapsSettings.boltzServiceFee}
                  % + on-chain fees
                </span>
              </div>
            </div>
          </CardContent>
          <CardFooter className="flex justify-end">
            <Button onClick={onDeactivate} disabled={loading} variant="outline">
              <XCircleIcon className="h-4 w-4 mr-2" />
              Deactivate
            </Button>
          </CardFooter>
        </Card>
      )}
    </>
  );
}

export default Swaps;
