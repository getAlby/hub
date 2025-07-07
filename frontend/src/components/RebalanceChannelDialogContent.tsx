import { AlertTriangleIcon, ExternalLinkIcon } from "lucide-react";
import React from "react";
import ExternalLink from "src/components/ExternalLink";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useChannels } from "src/hooks/useChannels";
import { request } from "src/utils/request";
import {
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";

type Props = {
  receiveThroughNodePubkey: string;
};

export function RebalanceChannelDialogContent({
  receiveThroughNodePubkey,
}: Props) {
  const [amount, setAmount] = React.useState("");
  const { toast } = useToast();
  const { data: channels, mutate: reloadChannels } = useChannels();
  const [isRebalancing, setRebalancing] = React.useState(false);

  const inputRef = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    setTimeout(() => {
      // for some reason `autoFocus` is not working on this input
      inputRef.current?.focus();
    }, 100);
  }, [inputRef]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setRebalancing(true);
    try {
      if (!channels) {
        throw new Error("channels not loaded");
      }

      if (
        channels
          .filter(
            (channel) => channel.remotePubkey === receiveThroughNodePubkey
          )
          .every((channel) => channel.remoteBalance / 1000 < parseInt(amount))
      ) {
        throw new Error("not enough receiving capacity in this channel");
      }

      if (
        channels
          .filter(
            (channel) => channel.remotePubkey === receiveThroughNodePubkey
          )
          .every(
            (channel) => channel.localSpendableBalance / 1000 > parseInt(amount)
          )
      ) {
        throw new Error(
          "cannot swap a smaller amount than your current channel balance"
        );
      }

      const response = await request<{ totalFeeSat: number }>(
        `/api/channels/rebalance`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            receiveThroughNodePubkey,
            amountSat: parseInt(amount),
          }),
        }
      );
      if (!response) {
        throw new Error("No rebalance response received");
      }

      await reloadChannels();
      toast({
        title:
          "Successfully rebalanced channels. Total fee: " +
          response.totalFeeSat +
          " sats",
      });
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        description: "Something went wrong: " + error,
      });
    }
    setRebalancing(false);
  }

  return (
    <AlertDialogContent>
      <form onSubmit={handleSubmit}>
        <AlertDialogHeader>
          <AlertDialogTitle>Rebalance In</AlertDialogTitle>
          <AlertDialogDescription>
            <p className="mb-4">
              Rebalance funds from other channels into this channel.
            </p>
            <Label htmlFor="fee" className="block mb-2">
              Rebalance amount (sats)
            </Label>
            <Input
              ref={inputRef}
              id="amount"
              name="amount"
              type="number"
              required
              autoFocus
              min={10000}
              value={amount}
              onChange={(e) => {
                setAmount(e.target.value.trim());
              }}
            />
            <p className="mt-2 text-xs text-muted-foreground">
              Fee: ~0.2%
              {!!amount && (
                <> ({Math.floor(parseInt(amount || "0") * 0.01)} sats)</>
              )}
            </p>
            <ExternalLink
              to="https://guides.getalby.com/user-guide/alby-hub/faq/can-i-rebalance-funds-from-one-of-my-channels-to-another"
              className="underline flex items-center mt-4"
            >
              Learn more about rebalancing between channels
              <ExternalLinkIcon className="w-4 h-4 ml-2" />
            </ExternalLink>
            <Alert className="mt-2">
              <AlertTriangleIcon className="h-4 w-4" />
              <AlertTitle>Rebalancing is in beta</AlertTitle>
              <AlertDescription>
                Funds may be rebalanced out of unexpected channels.
              </AlertDescription>
            </Alert>
            {channels &&
              channels.filter(
                (channel) => channel.remotePubkey === receiveThroughNodePubkey
              ).length > 1 && (
                <Alert className="mt-2">
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>
                    Multiple channels with same counterparty
                  </AlertTitle>
                  <AlertDescription>
                    Funds may be rebalanced to an unexpected channel.
                  </AlertDescription>
                </Alert>
              )}
            {channels?.some(
              (channel) =>
                channel.remotePubkey !== receiveThroughNodePubkey &&
                channel.localSpendableBalance / 1000 <
                  (channels.find(
                    (other) => other.remotePubkey === receiveThroughNodePubkey
                  )?.localBalance || 0)
            ) && (
              <Alert className="mt-2">
                <AlertTriangleIcon className="h-4 w-4" />
                <AlertTitle>
                  You have another channel with less funds
                </AlertTitle>
                <AlertDescription>
                  Consider choosing a channel with less spending balance to
                  rebalance into.
                </AlertDescription>
              </Alert>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="mt-4">
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <LoadingButton loading={isRebalancing}>Confirm</LoadingButton>
        </AlertDialogFooter>
      </form>
    </AlertDialogContent>
  );
}
