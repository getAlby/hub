import { AlertTriangleIcon } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { Channel } from "src/types";

type ChannelWarningProps = {
  channel: Channel;
};

export function ChannelWarning({ channel }: ChannelWarningProps) {
  const capacity = channel.localBalance + channel.remoteBalance;
  let channelWarning = channel.error;
  if (!channelWarning && channel.status === "opening") {
    channelWarning = `Channel is currently being opened (${channel.confirmations} of ${channel.confirmationsRequired} confirmations). Once the required confirmation are reached, you will be able to send and receive on this channel.`;
  }
  if (!channelWarning && channel.status === "offline") {
    channelWarning =
      "This channel is currently offline and cannot be used to send or receive payments.";
  }
  if (!channelWarning && channel.localSpendableBalance > capacity * 0.9) {
    channelWarning =
      "Receiving capacity low. You may have trouble receiving payments through this channel.";
  }

  if (!channelWarning) {
    return null;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <AlertTriangleIcon className="w-4 h-4" />
        </TooltipTrigger>
        <TooltipContent className="max-w-[400px]">
          {channelWarning}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
