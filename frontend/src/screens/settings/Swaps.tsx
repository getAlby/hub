import { useState } from "react";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useToast } from "src/components/ui/use-toast";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { request } from "src/utils/request";

function Swaps() {
  const { toast } = useToast();
  const { data: onchainAddress } = useOnchainAddress();

  const [swapTo, setSwapTo] = useState("hub");
  const [balanceThreshold, setBalanceThreshold] = useState("");
  const [swapAmount, setSwapAmount] = useState("");
  const [destination, setDestination] = useState(onchainAddress);
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/settings/swaps", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          swapAmount,
          balanceThreshold: parseInt(balanceThreshold),
          destination: swapTo === "hub" ? onchainAddress : destination,
        }),
      });
      toast({ title: "Saved successfully." });
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

  if (!onchainAddress) {
    return <Loading />;
  }

  return (
    <>
      <SettingsHeader
        title="Swaps"
        description="Swap bitcoin lightning into your on-chain balance"
      />
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <div className="grid gap-2">
          <Label>Spending balance threshold</Label>
          <Input
            type="number"
            placeholder="Swap out as soon as this amount is reached"
            value={balanceThreshold}
            onChange={(e) => setBalanceThreshold(e.target.value)}
          />
        </div>
        <div className="grid gap-2">
          <Label>Swap amount</Label>
          <Input
            type="number"
            placeholder="How much do you want to swap out?"
            value={swapAmount}
            min={50000}
            onChange={(e) => setSwapAmount(e.target.value)}
          />
        </div>
        <Label>Swap to</Label>
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
          <div className="grid gap-2">
            <Label>Receiving on-chain address</Label>
            <Input
              placeholder="bc1..."
              value={destination}
              onChange={(e) => setDestination(e.target.value)}
            />
          </div>
        )}
        <div>
          <LoadingButton
            loading={loading}
            disabled={!destination || !balanceThreshold}
          >
            Enable Auto Swaps
          </LoadingButton>
        </div>
      </form>
    </>
  );
}

export default Swaps;
