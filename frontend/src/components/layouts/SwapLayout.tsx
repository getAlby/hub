import { XCircleIcon } from "lucide-react";
import { useState } from "react";
import { Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useSwaps } from "src/hooks/useSwaps";
import { SwapsSettingsResponse } from "src/types";
import { request } from "src/utils/request";

export default function SwapLayout() {
  const { data: swapsSettings } = useSwaps();

  if (!swapsSettings) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader title="Swap" />
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          <Outlet />
        </div>
        {swapsSettings.enabled && <ActiveSwaps swaps={swapsSettings} />}
      </div>
    </div>
  );
}

function ActiveSwaps({ swaps }: { swaps: SwapsSettingsResponse }) {
  const { toast } = useToast();
  const { mutate } = useSwaps();

  const [loading, setLoading] = useState(false);

  const onDeactivate = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/wallet/swaps", {
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
    <div className="border-t-2 lg:border-t-0 pt-8 lg:pt-0">
      <Card>
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
                {swaps.destination}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="font-medium">Spending Balance Threshold</span>
              <span className="text-muted-foreground text-right">
                {new Intl.NumberFormat().format(swaps.balanceThreshold)} sats
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="font-medium">Swap amount</span>
              <span className="text-muted-foreground text-right">
                {new Intl.NumberFormat().format(swaps.swapAmount)} sats
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="font-medium">Fee</span>
              <span className="text-muted-foreground text-right">
                {swaps.albyServiceFee + swaps.boltzServiceFee}% + on-chain fees
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
    </div>
  );
}
