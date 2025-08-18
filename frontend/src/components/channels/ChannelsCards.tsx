import { InfoIcon } from "lucide-react";
import { ChannelDropdownMenu } from "src/components/channels/ChannelDropdownMenu";
import { ChannelWarning } from "src/components/channels/ChannelWarning";
import { Badge } from "src/components/ui/badge.tsx";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress.tsx";
import { Separator } from "src/components/ui/separator";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { formatAmount } from "src/lib/utils.ts";
import { Channel, LongUnconfirmedZeroConfChannel } from "src/types";

type ChannelsCardsProps = {
  channels?: Channel[];
  longUnconfirmedZeroConfChannels: LongUnconfirmedZeroConfChannel[];
};

export function ChannelsCards({
  channels,
  longUnconfirmedZeroConfChannels,
}: ChannelsCardsProps) {
  if (!channels?.length) {
    return null;
  }

  return (
    <Card className="lg:hidden">
      <CardHeader className="w-full pb-2 text-2xl font-semibold">
        Channels
      </CardHeader>
      <CardContent>
        {channels
          .sort((a, b) =>
            a.localBalance + a.remoteBalance > b.localBalance + b.remoteBalance
              ? -1
              : 1
          )
          .map((channel, index) => {
            const unconfirmedChannel = longUnconfirmedZeroConfChannels.find(
              (uc) => uc.id === channel.id
            );
            return (
              <ChannelCard
                key={index}
                addSeparator={index > 0}
                channel={channel}
                unconfirmedChannel={unconfirmedChannel}
              />
            );
          })}
      </CardContent>
    </Card>
  );
}
type ChannelCardProps = {
  channel: Channel;
  unconfirmedChannel: LongUnconfirmedZeroConfChannel | undefined;
  addSeparator: boolean;
};

function ChannelCard({
  channel,
  unconfirmedChannel,
  addSeparator,
}: ChannelCardProps) {
  const { data: node } = useNodeDetails(channel.remotePubkey);
  const alias = node?.alias || "Unknown";
  const capacity = channel.localBalance + channel.remoteBalance;

  return (
    <>
      {addSeparator && <Separator className="mt-6 -mb-2" />}
      <div className="flex flex-col items-start w-full">
        <CardHeader className="pb-4 px-0 w-full">
          <CardTitle className="w-full">
            <div className="flex items-center justify-between">
              <div className="flex-1 whitespace-nowrap text-ellipsis font-semibold truncate leading-normal">
                {alias}
              </div>
              <ChannelDropdownMenu alias={alias} channel={channel} />
            </div>
          </CardTitle>
        </CardHeader>
        <CardDescription className="w-full flex flex-col gap-4">
          <div className="flex w-full justify-between items-center">
            <p className="text-muted-foreground font-medium">Status</p>
            {channel.status == "online" ? (
              unconfirmedChannel ? (
                <Badge variant="outline" title={unconfirmedChannel.message}>
                  Unconfirmed
                </Badge>
              ) : (
                <Badge variant="positive">Online</Badge>
              )
            ) : channel.status == "opening" ? (
              <Badge variant="outline">Opening</Badge>
            ) : (
              <Badge variant="warning">Offline</Badge>
            )}
          </div>
          <div className="flex w-full justify-between items-center">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className="flex flex-row gap-1 items-center text-muted-foreground">
                    Type
                    <InfoIcon className="h-3 w-3 shrink-0" />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  The type of lightning channel, By default private channel is
                  recommended. If you a podcaster or musician and expect to
                  receive keysend or Value4Value payments you will need a public
                  channel.
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <p className="text-foreground">
              {channel.public ? "Public" : "Private"}
            </p>
          </div>
          <div className="flex justify-between items-center">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className="flex flex-row gap-1 items-center text-muted-foreground">
                    Capacity
                    <InfoIcon className="h-3 w-3 shrink-0" />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  Total Spending and Receiving capacity of your lightning
                  channel.
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <p className="text-foreground">{formatAmount(capacity)} sats</p>
          </div>
          <div className="flex justify-between items-center">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className="flex flex-row gap-1 items-center text-muted-foreground">
                    Reserve
                    <InfoIcon className="h-3 w-3 shrink-0" />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  Funds each participant sets aside to discourage cheating by
                  ensuring each party has something at stake. This reserve
                  cannot be spent during the channel's lifetime and typically
                  amounts to 1% of the channel capacity. The reserve will
                  automatically be filled when payments are received on the
                  channel.
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <p className="text-foreground">
              {channel.localBalance <
                channel.unspendablePunishmentReserve * 1000 && (
                <>
                  {formatAmount(
                    Math.min(
                      channel.localBalance,
                      channel.unspendablePunishmentReserve * 1000
                    )
                  )}{" "}
                  /{" "}
                </>
              )}
              {formatAmount(channel.unspendablePunishmentReserve * 1000)} sats
            </p>
          </div>
          <div className="flex justify-between items-center">
            <p className="text-muted-foreground font-medium text-sm">
              Spending
            </p>
            <p className="text-muted-foreground font-medium text-sm">
              Receiving
            </p>
          </div>
          <div className="flex gap-2 items-center">
            <div className="flex-1 relative">
              <Progress
                value={(channel.localSpendableBalance / capacity) * 100}
                className="h-6 absolute"
              />
              <div className="flex flex-row w-full justify-between px-2 text-xs items-center h-6 mix-blend-exclusion text-white">
                <span title={channel.localSpendableBalance / 1000 + " sats"}>
                  {formatAmount(channel.localSpendableBalance)} sats
                </span>
                <span title={channel.remoteBalance / 1000 + " sats"}>
                  {formatAmount(channel.remoteBalance)} sats
                </span>
              </div>
            </div>
            <ChannelWarning channel={channel} />
          </div>
        </CardDescription>
      </div>
    </>
  );
}
