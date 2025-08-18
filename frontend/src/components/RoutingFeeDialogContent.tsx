import { ExternalLinkIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import ExternalLink from "src/components/ExternalLink";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useChannels } from "src/hooks/useChannels";
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

export function RoutingFeeDialogContent({ channel }: Props) {
  const currentFee: number = React.useMemo(() => {
    return Math.floor(channel.forwardingFeeBaseMsat / 1000);
  }, [channel.forwardingFeeBaseMsat]);
  const [forwardingFee, setForwardingFee] = React.useState(
    currentFee ? currentFee.toString() : ""
  );
  const { mutate: reloadChannels } = useChannels();

  async function updateFee() {
    try {
      const forwardingFeeBaseMsat = +forwardingFee * 1000;

      console.info(
        `🎬 Updating channel ${channel.id} with ${channel.remotePubkey}`
      );

      await request(
        `/api/peers/${channel.remotePubkey}/channels/${channel.id}`,
        {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            forwardingFeeBaseMsat: forwardingFeeBaseMsat,
          } as UpdateChannelRequest),
        }
      );

      await reloadChannels();
      toast("Successfully updated channel");
    } catch (error) {
      console.error(error);
      toast.error("Something went wrong", {
        description: "" + error,
      });
    }
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Update Channel Routing Fee</AlertDialogTitle>
        <AlertDialogDescription>
          <p className="mb-4">
            Adjust the fee you charge for each payment routed through this
            channel. A high fee (e.g. 100,000 sats) can be set to prevent
            unwanted routing. No matter the fee, you can still receive payments.{" "}
          </p>
          <Label htmlFor="fee" className="block mb-2">
            Routing Fee (sats)
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
            to="https://guides.getalby.com/user-guide/alby-hub/faq/how-can-i-change-routing-fees#understanding-routing-fees-and-alby-hub"
            className="underline flex items-center mt-4"
          >
            Learn more about routing fees
            <ExternalLinkIcon className="size-4 ml-2" />
          </ExternalLink>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction
          disabled={(parseInt(forwardingFee) || 0) == currentFee}
          onClick={updateFee}
        >
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}
