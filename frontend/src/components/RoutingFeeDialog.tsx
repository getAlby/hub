import { ExternalLinkIcon } from "lucide-react";
import React from "react";
import ExternalLink from "src/components/ExternalLink";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { toast } from "src/components/ui/use-toast";
import { useChannels } from "src/hooks/useChannels";
import { useCSRF } from "src/hooks/useCSRF";
import { Channel, UpdateChannelRequest } from "src/types";
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
  channel: Channel;
};

export function RoutingFeeDialog({ channel }: Props) {
  const currentFee: number = React.useMemo(() => {
    return Math.floor(channel.forwardingFeeBaseMsat / 1000);
  }, [channel.forwardingFeeBaseMsat]);
  const [forwardingFee, setForwardingFee] = React.useState(
    currentFee ? currentFee.toString() : ""
  );
  const { mutate: reloadChannels } = useChannels();
  const { data: csrf } = useCSRF();

  async function updateFee() {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }

      const forwardingFeeBaseMsat = +forwardingFee * 1000;

      console.info(
        `🎬 Updating channel ${channel.id} with ${channel.remotePubkey}`
      );

      await request(
        `/api/peers/${channel.remotePubkey}/channels/${channel.id}`,
        {
          method: "PATCH",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            forwardingFeeBaseMsat: forwardingFeeBaseMsat,
          } as UpdateChannelRequest),
        }
      );

      await reloadChannels();
      toast({ title: "Sucessfully updated channel" });
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
        <AlertDialogTitle>Update Channel Forwarding Fee</AlertDialogTitle>
        <AlertDialogDescription>
          <p className="mb-4">
            Adjust the fee you charge for each payment passing through your
            channel.{" "}
            <span className="text-primary font-medium">Current fee:</span>{" "}
            {currentFee} sats
          </p>
          <Label htmlFor="fee" className="block mb-2">
            Base Forwarding Fee (sats)
          </Label>
          <Input
            id="fee"
            name="fee"
            type="number"
            required
            autoFocus
            min={0}
            value={forwardingFee}
            onChange={(e) => {
              setForwardingFee(e.target.value.trim());
            }}
          />
          <ExternalLink
            to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/faq-alby-hub/how-can-i-change-routing-fees#understanding-routing-fees-and-alby-hub"
            className="underline flex items-center mt-4"
          >
            Learn more about routing fees
            <ExternalLinkIcon className="w-4 h-4 ml-2" />
          </ExternalLink>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction onClick={updateFee}>Confirm</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}
