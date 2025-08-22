import { ExternalLinkIcon } from "lucide-react";
import React from "react";
import ExternalLink from "src/components/ExternalLink";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
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
  const currentBaseFeeSats: number = Math.floor(
    channel.forwardingFeeBaseMsat / 1000
  );
  const currentFeePPM: number = channel.forwardingFeeProportionalMillionths;

  const [baseFeeSats, setBaseFeeSats] = React.useState(
    currentBaseFeeSats !== undefined ? currentBaseFeeSats.toString() : ""
  );
  const [
    forwardingFeeProportionalMillionths,
    setForwardingFeeProportionalMillionths,
  ] = React.useState(
    currentFeePPM !== undefined ? currentFeePPM.toString() : ""
  );
  const { toast } = useToast();
  const { mutate: reloadChannels } = useChannels();

  async function updateFee() {
    try {
      const forwardingFeeBaseMsat = +baseFeeSats * 1000;

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
            forwardingFeeProportionalMillionths:
              +forwardingFeeProportionalMillionths,
          } as UpdateChannelRequest),
        }
      );

      await reloadChannels();
      toast({ title: "Successfully updated channel" });
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
        <AlertDialogTitle>Update Channel Routing Fee</AlertDialogTitle>
        <AlertDialogDescription>
          <p className="mb-4">
            Adjust the fee you charge for each payment routed through this
            channel. A high fee (e.g. 100,000 sats) can be set to prevent
            unwanted routing. No matter the fee, you can still receive payments.{" "}
          </p>
          <Label htmlFor="fee" className="block mb-2">
            Base Routing Fee (sats)
          </Label>
          <Input
            id="fee"
            name="fee"
            type="number"
            required
            autoFocus
            min={0}
            value={baseFeeSats}
            onChange={(e) => {
              setBaseFeeSats(e.target.value.trim());
            }}
          />
          <Label htmlFor="fee" className="block mt-4 mb-2">
            PPM Fee (1 PPM = 1 per 1 million sats)
          </Label>
          <Input
            id="fee"
            name="fee"
            type="number"
            required
            autoFocus
            min={0}
            value={forwardingFeeProportionalMillionths}
            onChange={(e) => {
              setForwardingFeeProportionalMillionths(e.target.value.trim());
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
          disabled={
            (parseInt(baseFeeSats) || 0) === currentBaseFeeSats &&
            (parseInt(forwardingFeeProportionalMillionths) || 0) ===
              currentFeePPM
          }
          onClick={updateFee}
        >
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}
