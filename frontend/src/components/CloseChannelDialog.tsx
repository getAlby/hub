import { AlertTriangleIcon, CopyIcon, ExternalLinkIcon } from "lucide-react";
import React from "react";
import ExternalLink from "src/components/ExternalLink";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { toast } from "src/components/ui/use-toast";
import { useChannels } from "src/hooks/useChannels";
import { useCSRF } from "src/hooks/useCSRF";
import { copyToClipboard } from "src/lib/clipboard";
import { Channel, CloseChannelResponse } from "src/types";
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
  alias: string;
  channel: Channel;
};

export function CloseChannelDialog({ alias, channel }: Props) {
  const [closeType, setCloseType] = React.useState("normal");
  const [step, setStep] = React.useState(channel.active ? 2 : 1);
  const [fundingTxId, setFundingTxId] = React.useState("");
  const { data: channels, mutate: reloadChannels } = useChannels();
  const { data: csrf } = useCSRF();

  const onContinue = () => {
    setStep(step + 1);
  };

  const copy = (text: string) => {
    copyToClipboard(text);
    toast({ title: "Copied to clipboard." });
  };

  async function closeChannel() {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }

      console.info(`🎬 Closing channel with ${channel.remotePubkey}`);

      const closeChannelResponse = await request<CloseChannelResponse>(
        `/api/peers/${channel.remotePubkey}/channels/${channel.id}?force=${
          closeType === "force"
        }`,
        {
          method: "DELETE",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
        }
      );

      if (!closeChannelResponse) {
        throw new Error("Error closing channel");
      }

      const closedChannel = channels?.find(
        (c) => c.id === channel.id && c.remotePubkey === channel.remotePubkey
      );
      console.info("Closed channel", closedChannel);
      if (closedChannel) {
        setFundingTxId(closedChannel.fundingTxId);
        setStep(step + 1);
      }
      toast({ title: "Sucessfully closed channel" });
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
      {step === 1 && (
        <>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Are you sure you want to close the channel with {alias}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              This channel is inactive. Some channels require up to 6 onchain
              confirmations before they are usable. Proceed only if you still
              want to continue
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <Button onClick={onContinue}>Continue</Button>
          </AlertDialogFooter>
        </>
      )}

      {step === 2 && (
        <>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Are you sure you want to close the channel with {alias}?
            </AlertDialogTitle>
            <AlertDialogDescription className="text-left">
              <div>
                <p className="text-primary font-medium">Node ID</p>
                <p className="break-all">{channel.remotePubkey}</p>
              </div>
              <div className="mt-4">
                <p className="text-primary font-medium">Channel ID</p>
                <p className="break-all">{channel.id}</p>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <Button onClick={onContinue}>Continue</Button>
          </AlertDialogFooter>
        </>
      )}

      {step === 3 && (
        <>
          <AlertDialogHeader>
            <AlertDialogTitle>Select mode of channel closure</AlertDialogTitle>
            <AlertDialogDescription className="text-left">
              {closeType === "force" && (
                <Alert className="mb-4">
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>Heads up!</AlertTitle>
                  <AlertDescription>
                    Your channel balance will be locked for up to two weeks if
                    you force close
                  </AlertDescription>
                </Alert>
              )}
              <div className="flex flex-col gap-4 text-xs mt-2">
                <div className="items-top flex space-x-2">
                  <Checkbox
                    id="normal"
                    onCheckedChange={() =>
                      setCloseType(closeType === "normal" ? "force" : "normal")
                    }
                    checked={closeType === "normal"}
                  />
                  <div className="grid gap-1.5 leading-none">
                    <label
                      htmlFor="normal"
                      className="text-primary text-sm font-medium leading-none cursor-pointer"
                    >
                      Normal Close
                    </label>
                    <p className="text-sm text-muted-foreground">
                      Closes the channel cooperatively, usually faster and with
                      lower fees
                    </p>
                  </div>
                </div>
                <div className="items-top flex space-x-2">
                  <Checkbox
                    id="force"
                    onCheckedChange={() =>
                      setCloseType(closeType === "force" ? "normal" : "force")
                    }
                    checked={closeType === "force"}
                  />
                  <div className="grid gap-1.5 leading-none">
                    <label
                      htmlFor="force"
                      className="text-primary text-sm font-medium leading-none cursor-pointer"
                    >
                      Force Close
                    </label>
                    <p className="text-sm text-muted-foreground">
                      Closes the channel unilaterally, can take longer and might
                      incur higher fees
                    </p>
                  </div>
                </div>
              </div>
              <ExternalLink
                to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/faq-alby-hub/how-can-i-close-this-channel-what-happens-to-the-sats-in-this-channel"
                className="underline flex items-center mt-4"
              >
                Learn more about closing channels
                <ExternalLinkIcon className="w-4 h-4 ml-2" />
              </ExternalLink>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <Button onClick={closeChannel}>Close Channel</Button>
          </AlertDialogFooter>
        </>
      )}

      {step === 4 && (
        <>
          <AlertDialogHeader>
            <AlertDialogTitle>Channel closed successfully</AlertDialogTitle>
            <AlertDialogDescription className="text-left">
              <p className="text-primary font-medium">Funding Transaction Id</p>
              <div className="flex items-center justify-between gap-4">
                <p className="break-all">{fundingTxId}</p>
                <CopyIcon
                  className="cursor-pointer text-muted-foreground w-4 h-4"
                  onClick={() => {
                    copy(fundingTxId);
                  }}
                />
              </div>
              <ExternalLink
                to={`https://mempool.space/tx/${fundingTxId}`}
                className="underline flex items-center mt-2"
              >
                View on Mempool
                <ExternalLinkIcon className="w-4 h-4 ml-2" />
              </ExternalLink>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={async () => {
                await reloadChannels();
              }}
            >
              Done
            </AlertDialogCancel>
          </AlertDialogFooter>
        </>
      )}
    </AlertDialogContent>
  );
}
