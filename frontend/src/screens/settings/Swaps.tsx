import { useState } from "react";
import SettingsHeader from "src/components/SettingsHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { request } from "src/utils/request";

function Swaps() {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);

  const [balanceThreshold, setBalanceThreshold] = useState("");
  const [destination, setDestination] = useState("");
  const [swapAmount, setSwapAmount] = useState("");

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
          balanceThreshold: parseInt(balanceThreshold),
          swapAmount: parseInt(swapAmount),
          destination,
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

  return (
    <>
      <SettingsHeader title="Swaps" description={""} />
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <div className="grid gap-1.5">
          <Label>Spending balance threshold</Label>
          <Input
            placeholder="Swap out as soon as this amount is reached"
            value={balanceThreshold}
            onChange={(e) => setBalanceThreshold(e.target.value)}
          />
        </div>
        <div className="grid gap-1.5">
          <Label>Swap amount</Label>
          <Input
            placeholder="How much do you want to swap out?"
            value={swapAmount}
            onChange={(e) => setSwapAmount(e.target.value)}
          />
        </div>
        <div className="grid gap-1.5">
          <Label>Bitcoin address</Label>
          <Input
            placeholder="bc1..."
            value={destination}
            onChange={(e) => setDestination(e.target.value)}
          />
        </div>
        <div>
          <LoadingButton loading={loading}>Save</LoadingButton>
        </div>
      </form>
    </>
  );
}

export default Swaps;
