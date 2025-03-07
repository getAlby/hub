import React from "react";
import { Link } from "react-router-dom";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast.ts";
import { useChannels } from "src/hooks/useChannels";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { Channel, CreateInvoiceRequest, Transaction } from "src/types";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request";

type SwapDialogsProps = {
  swapInDialogOpen: boolean;
  setSwapInDialogOpen: (open: boolean) => void;
  swapOutDialogOpen: boolean;
  setSwapOutDialogOpen: (open: boolean) => void;
};

export function SwapDialogs({
  swapInDialogOpen,
  setSwapInDialogOpen,
  swapOutDialogOpen,
  setSwapOutDialogOpen,
}: SwapDialogsProps) {
  const { toast } = useToast();
  const [swapInAmount, setSwapInAmount] = React.useState("");
  const [swapOutAmount, setSwapOutAmount] = React.useState("");

  const [loadingSwap, setLoadingSwap] = React.useState(false);
  const { getNewAddress } = useOnchainAddress();
  const { data: channels } = useChannels();

  const findChannelWithLargestBalance = React.useCallback(
    (
      balanceType: "remoteBalance" | "localSpendableBalance"
    ): Channel | undefined => {
      if (!channels || channels.length === 0) {
        return undefined;
      }

      return channels.reduce((prevLargest, current) => {
        return current[balanceType] > prevLargest[balanceType]
          ? current
          : prevLargest;
      }, channels[0]);
    },
    [channels]
  );

  React.useEffect(() => {
    if (swapOutDialogOpen) {
      setSwapOutAmount(
        Math.floor(
          ((findChannelWithLargestBalance("localSpendableBalance")
            ?.localSpendableBalance || 0) *
            0.9) /
            1000
        ).toString()
      );
    }
  }, [findChannelWithLargestBalance, swapOutDialogOpen]);

  React.useEffect(() => {
    if (swapInDialogOpen) {
      setSwapInAmount(
        Math.floor(
          ((findChannelWithLargestBalance("remoteBalance")?.remoteBalance ||
            0) *
            0.9) /
            1000
        ).toString()
      );
    }
  }, [findChannelWithLargestBalance, swapInDialogOpen]);

  return (
    <>
      <Dialog open={swapOutDialogOpen} onOpenChange={setSwapOutDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Swap out funds</DialogTitle>
            <DialogDescription>
              Funds from one of your channels will be sent to your on-chain
              balance via a swap service. This helps restore your inbound
              liquidity.
            </DialogDescription>
          </DialogHeader>
          <div className="grid grid-cols-4">
            <Label className="pt-3">Amount (sats)</Label>
            <div className="col-span-3">
              <Input
                value={swapOutAmount}
                onChange={(e) => setSwapOutAmount(e.target.value)}
              />
              <p className="text-muted-foreground text-xs p-2">
                The amount is set to 90% of the maximum spending capacity
                available in one of your lightning channels.
              </p>
            </div>
          </div>
          <DialogFooter>
            <LoadingButton
              loading={loadingSwap}
              type="submit"
              onClick={async () => {
                setLoadingSwap(true);
                const onchainAddress = await getNewAddress();
                if (onchainAddress) {
                  openLink(
                    `https://boltz.exchange/?sendAsset=LN&receiveAsset=BTC&sendAmount=${swapOutAmount}&destination=${onchainAddress}&ref=alby`
                  );
                  setSwapOutDialogOpen(false);
                }
                setLoadingSwap(false);
              }}
            >
              Swap out
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={swapInDialogOpen} onOpenChange={setSwapInDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Swap in funds</DialogTitle>
            <DialogDescription>
              Swap on-chain funds into your lightning channels via a swap
              service, increasing your spending balance using on-chain funds
              from{" "}
              <Link to="/wallet/withdraw" className="underline">
                your hub
              </Link>{" "}
              or an external wallet.
            </DialogDescription>
          </DialogHeader>
          <div className="grid grid-cols-4">
            <Label className="pt-3">Amount (sats)</Label>
            <div className="col-span-3">
              <Input
                value={swapInAmount}
                onChange={(e) => setSwapInAmount(e.target.value)}
              />
              <p className="text-muted-foreground text-xs p-2">
                The amount is set to 90% of the maximum receiving capacity
                available in one of your lightning channels.
              </p>
            </div>
          </div>
          <DialogFooter>
            <LoadingButton
              loading={loadingSwap}
              type="submit"
              onClick={async () => {
                setLoadingSwap(true);
                try {
                  const transaction = await request<Transaction>(
                    "/api/invoices",
                    {
                      method: "POST",
                      headers: {
                        "Content-Type": "application/json",
                      },
                      body: JSON.stringify({
                        amount: parseInt(swapInAmount) * 1000,
                        description: "Boltz Swap In",
                      } as CreateInvoiceRequest),
                    }
                  );
                  if (!transaction) {
                    throw new Error("no transaction in response");
                  }
                  openLink(
                    `https://boltz.exchange/?sendAsset=BTC&receiveAsset=LN&sendAmount=${swapInAmount}&destination=${transaction.invoice}&ref=alby`
                  );
                  setSwapInDialogOpen(false);
                } catch (error) {
                  toast({
                    variant: "destructive",
                    title: "Failed to generate swap invoice",
                    description: "" + error,
                  });
                }
                setLoadingSwap(false);
              }}
            >
              Swap in
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
