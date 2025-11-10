import {
  AlertCircleIcon,
  AlertTriangleIcon,
  CopyIcon,
  ExternalLinkIcon,
} from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { SwapAlert } from "src/components/channels/SwapAlert";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { MempoolAlert } from "src/components/MempoolAlert";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
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

export function CloseChannelDialogContent({ alias, channel }: Props) {
  const [closeType, setCloseType] = React.useState("normal");
  const [step, setStep] = React.useState(channel.active ? 2 : 1);
  const [fundingTxId, setFundingTxId] = React.useState("");
  const { data: info } = useInfo();
  const { mutate: reloadBalances } = useBalances();
  const { data: channels, mutate: reloadChannels } = useChannels();

  const onContinue = () => {
    setStep(step + 1);
  };

  const copy = (text: string) => {
    copyToClipboard(text);
  };

  async function closeChannel() {
    try {
      console.info(`🎬 Closing channel with ${channel.remotePubkey}`);

      const closeChannelResponse = await request<CloseChannelResponse>(
        `/api/peers/${channel.remotePubkey}/channels/${channel.id}?force=${
          closeType === "force"
        }`,
        {
          method: "DELETE",
          headers: {
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
      toast("Successfully closed channel");
    } catch (error) {
      console.error(error);
      toast.error("Something went wrong: " + error);
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
              confirmations before they are usable.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <Button onClick={onContinue}>Confirm</Button>
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
              <SwapAlert minChannels={0} className="mb-4" />
              <Alert className="mb-4">
                <AlertCircleIcon className="h-4 w-4" />
                <AlertDescription>
                  <div>
                    Closing this channel will move{" "}
                    <FormattedBitcoinAmount amount={channel.localBalance} /> in
                    this channel to your on-chain balance and reduce your
                    receive limit by{" "}
                    <FormattedBitcoinAmount amount={channel.remoteBalance} />.
                  </div>
                </AlertDescription>
              </Alert>
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
              <div className="mb-4">
                <MempoolAlert />
              </div>
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
              <RadioGroup
                defaultValue="normal"
                value={closeType}
                onValueChange={() =>
                  setCloseType(closeType === "normal" ? "force" : "normal")
                }
                className="mt-2"
              >
                <div className="flex items-start space-x-2 mb-2">
                  <RadioGroupItem
                    value="normal"
                    id="normal"
                    className="shrink-0"
                  />
                  <div className="grid gap-1.5">
                    <Label
                      htmlFor="normal"
                      className="text-primary font-medium cursor-pointer"
                    >
                      Normal Close (Recommended)
                    </Label>
                    <p className="text-sm text-muted-foreground">
                      Attempt to settle with your channel partner for a quick,
                      low-cost closure. Your funds should be available on-chain
                      within an hour. If your partner is offline or agreement
                      cannot be met, a force closure will be initiated.
                    </p>
                  </div>
                </div>
                <div className="flex items-start space-x-2">
                  <RadioGroupItem
                    value="force"
                    id="force"
                    className="shrink-0"
                  />
                  <div className="grid gap-1.5">
                    <Label
                      htmlFor="force"
                      className="text-primary font-medium cursor-pointer"
                    >
                      Force Close
                    </Label>
                    <p className="text-sm text-muted-foreground">
                      You close the channel alone. Your funds may be locked for
                      up to two weeks and may incur higher fees. Only try this
                      if a normal closure does not work.
                    </p>
                  </div>
                </div>
              </RadioGroup>
              <ExternalLink
                to="https://guides.getalby.com/user-guide/alby-hub/faq/how-can-i-close-a-channel-what-happens-to-the-sats-in-this-channel"
                className="underline flex items-center mt-4"
              >
                Learn more about closing channels
                <ExternalLinkIcon className="size-4 ml-2" />
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
                  className="cursor-pointer text-muted-foreground size-4"
                  onClick={() => {
                    copy(fundingTxId);
                  }}
                />
              </div>
              <ExternalLink
                to={`${info?.mempoolUrl}/tx/${fundingTxId}`}
                className="underline flex items-center mt-2"
              >
                View on Mempool
                <ExternalLinkIcon className="size-4 ml-2" />
              </ExternalLink>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={async () => {
                await reloadChannels();
                await reloadBalances();
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
