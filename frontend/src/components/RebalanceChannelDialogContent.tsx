import { ExternalLinkIcon } from "lucide-react";
import React from "react";
import ExternalLink from "src/components/ExternalLink";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useChannels } from "src/hooks/useChannels";
import { request } from "src/utils/request";
import {
  AlertDialogAction,
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
  const { mutate: reloadChannels } = useChannels();

  async function confirmRebalance() {
    try {
      // TODO: check that channels in this counterparty have less than other counterparties

      // TODO: ensure the rebalance amount is bigger than the current amount
      // in the user's channels with this counterparty. Otherwise the funds will be sent
      // on the same channel and the user just loses fees

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
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Rebalance In</AlertDialogTitle>
        <AlertDialogDescription>
          <p className="mb-4">
            Rebalance funds from other counterparty channels into channels with
            this counterparty.
          </p>
          <Label htmlFor="fee" className="block mb-2">
            Rebalance amount (sats)
          </Label>
          <Input
            id="amount"
            name="amount"
            type="number"
            required
            autoFocus
            min={0}
            value={amount}
            onChange={(e) => {
              setAmount(e.target.value.trim());
            }}
          />
          <p className="mt-2 text-xs text-muted-foreground">
            Fee: ~0.1% ({Math.floor(parseInt(amount || "0") * 0.01)} sats)
          </p>
          <ExternalLink
            to="https://guides.getalby.com/user-guide/alby-hub/faq/can-i-rebalance-funds-from-one-of-my-channels-to-another"
            className="underline flex items-center mt-4"
          >
            Learn more about rebalancing between channels
            <ExternalLinkIcon className="w-4 h-4 ml-2" />
          </ExternalLink>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction
          disabled={parseInt(amount) === 0}
          onClick={confirmRebalance}
        >
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}
