import { InfoIcon } from "lucide-react";
import { ChannelWarning } from "src/components/channels/ChannelWarning";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading.tsx";
import { Badge } from "src/components/ui/badge.tsx";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress.tsx";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "src/components/ui/table.tsx";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip.tsx";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { Channel, LongUnconfirmedZeroConfChannel } from "src/types";
import { ChannelDropdownMenu } from "./ChannelDropdownMenu";

type ChannelsTableProps = {
  channels?: Channel[];
  longUnconfirmedZeroConfChannels: LongUnconfirmedZeroConfChannel[];
};

export function ChannelsTable({
  channels,
  longUnconfirmedZeroConfChannels,
}: ChannelsTableProps) {
  if (channels && !channels.length) {
    return null;
  }

  return (
    <Card className="hidden lg:block">
      <CardHeader>
        <CardTitle className="text-2xl">Channels</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[160px] text-muted-foreground">
                Peer
              </TableHead>
              <TableHead className="w-[80px] text-muted-foreground">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <div className="flex flex-row gap-1 items-center text-muted-foreground">
                        Type
                        <InfoIcon className="h-3 w-3 shrink-0" />
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      The type of lightning channel, By default private channel
                      is recommended. If you a podcaster or musician and expect
                      to receive keysend or Value4Value payments you will need a
                      public channel.
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </TableHead>
              <TableHead className="w-[80px] text-muted-foreground">
                Status
              </TableHead>
              <TableHead className="w-[128px] text-muted-foreground">
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
              </TableHead>
              <TableHead className="w-[128px]">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <div className="flex flex-row gap-1 items-center text-muted-foreground">
                        Reserve
                        <InfoIcon className="h-3 w-3 shrink-0" />
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      Funds each participant sets aside to discourage cheating
                      by ensuring each party has something at stake. This
                      reserve cannot be spent during the channel's lifetime and
                      typically amounts to 1% of the channel capacity. The
                      reserve will automatically be filled when payments are
                      received on the channel.
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </TableHead>
              <TableHead className="w-[350px]">
                <div className="flex flex-row justify-between items-center gap-2 text-muted-foreground">
                  <div>Spending</div>
                  <div>Receiving</div>
                </div>
              </TableHead>
              <TableHead className="w-px"></TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {channels && channels.length > 0 && (
              <>
                {channels
                  .sort((a, b) =>
                    a.localBalance + a.remoteBalance >
                    b.localBalance + b.remoteBalance
                      ? -1
                      : 1
                  )
                  .map((channel) => {
                    const unconfirmedChannel =
                      longUnconfirmedZeroConfChannels.find(
                        (uc) => uc.id === channel.id
                      );
                    return (
                      <ChannelTableRow
                        key={channel.id}
                        channel={channel}
                        unconfirmedChannel={unconfirmedChannel}
                        hasMultipleChannels={channels.length > 1}
                      />
                    );
                  })}
              </>
            )}
            {!channels && (
              <TableRow>
                <TableCell colSpan={6}>
                  <Loading className="m-2" />
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

type ChannelTableRowProps = {
  channel: Channel;
  unconfirmedChannel: LongUnconfirmedZeroConfChannel | undefined;
  hasMultipleChannels: boolean;
};

function ChannelTableRow({
  channel,
  unconfirmedChannel,
  hasMultipleChannels,
}: ChannelTableRowProps) {
  const { data: peerDetails } = useNodeDetails(channel.remotePubkey);
  const capacity = channel.localBalance + channel.remoteBalance;
  const alias = peerDetails?.alias || "Unknown";

  return (
    <TableRow key={channel.id} className="channel sensitive">
      <TableCell>
        <span className="font-semibold text-sm mr-2">{alias}</span>
      </TableCell>
      <TableCell>{channel.public ? "Public" : "Private"}</TableCell>
      <TableCell>
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
      </TableCell>
      <TableCell>
        <FormattedBitcoinAmount amount={capacity} />
      </TableCell>
      <TableCell title={channel.unspendablePunishmentReserve + " sats"}>
        {channel.localBalance < channel.unspendablePunishmentReserve * 1000 && (
          <>
            <FormattedBitcoinAmount
              amount={Math.min(
                channel.localBalance,
                channel.unspendablePunishmentReserve * 1000
              )}
              showSymbol={false}
            />{" "}
            /{" "}
          </>
        )}
        <FormattedBitcoinAmount
          amount={channel.unspendablePunishmentReserve * 1000}
        />
      </TableCell>
      <TableCell>
        <div className="relative">
          <Progress
            value={(channel.localSpendableBalance / capacity) * 100}
            className="h-6 absolute"
          />
          <div className="flex flex-row w-full justify-between px-2 text-xs items-center h-6 mix-blend-exclusion text-white ">
            <span>
              <FormattedBitcoinAmount amount={channel.localSpendableBalance} />
            </span>
            <span>
              <FormattedBitcoinAmount amount={channel.remoteBalance} />
            </span>
          </div>
        </div>
      </TableCell>
      <TableCell>
        <ChannelWarning channel={channel} />
      </TableCell>
      <TableCell>
        <ChannelDropdownMenu
          alias={alias}
          channel={channel}
          hasMultipleChannels={hasMultipleChannels}
        />
      </TableCell>
    </TableRow>
  );
}
