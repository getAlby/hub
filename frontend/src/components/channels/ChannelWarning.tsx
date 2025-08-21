import { AlertTriangleIcon } from "lucide-react";
import { Link } from "react-router-dom";
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
  let channelWarning = <> {channel.error} </>;
  if (!channel.error && channel.status === "opening") {
    channelWarning = (
      <>
        {`Channel is currently being opened (${channel.confirmations} of ${channel.confirmationsRequired} confirmations). Once the required confirmation are reached, you will be able to send and receive on this channel.`}{" "}
      </>
    );
  }
  // console.info(channel.error);
  if (!channel.error && channel.status === "offline") {
    channelWarning = (
      <>
        This channel is currently offline and cannot be used to send or receive
        payments.
        <Link to="https://guides.getalby.com/user-guide/alby-hub/faq/why-is-my-channel-offline-and-what-should-i-do-now">
          Learn more here.
        </Link>
      </>
    );
  }

  if (
    !channel.error &&
    channel.status === "online" &&
    channel.localSpendableBalance > capacity * 0.9
  ) {
    channelWarning = (
      <>
        Receiving capacity low. You may have trouble receiving payments through
        this channel.
      </>
    );
  }

  if (!channelWarning) {
    return null;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <AlertTriangleIcon className="size-4" />
        </TooltipTrigger>
        <TooltipContent>{channelWarning}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
