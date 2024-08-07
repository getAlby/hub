import { InfoIcon } from "lucide-react";
import { ChannelWarning } from "src/components/channels/ChannelWarning";
import Loading from "src/components/Loading.tsx";
import { Badge } from "src/components/ui/badge.tsx";
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
import { formatAmount } from "src/lib/utils.ts";
import { Channel, Node } from "src/types";
import { ChannelDropdownMenu } from "./ChannelDropdownMenu";

type ChannelsTableProps = {
  channels?: Channel[];
  nodes?: Node[];
  editChannel(channel: Channel): void;
};

export function ChannelsTable({
  channels,
  nodes,
  editChannel,
}: ChannelsTableProps) {
  if (channels && !channels.length) {
    return null;
  }

  return (
    <div className="border rounded-lg max-w-full overflow-y-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[80px]">Status</TableHead>
            <TableHead>Node</TableHead>
            <TableHead className="w-[150px]">Capacity</TableHead>
            <TableHead className="w-[150px]">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <div className="flex flex-row gap-2 items-center">
                      Reserve
                      <InfoIcon className="h-4 w-4 shrink-0" />
                    </div>
                  </TooltipTrigger>
                  <TooltipContent className="w-[400px]">
                    Funds each participant sets aside to discourage cheating by
                    ensuring each party has something at stake. This reserve
                    cannot be spent during the channel's lifetime and typically
                    amounts to 1% of the channel capacity.
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </TableHead>
            <TableHead className="w-[300px]">
              <div className="flex flex-row justify-between items-center gap-2">
                <div>Spending</div>
                <div>Receiving</div>
              </div>
            </TableHead>
            <TableHead className="w-[24px]"></TableHead>
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
                  const node = nodes?.find(
                    (n) => n.public_key === channel.remotePubkey
                  );
                  const alias = node?.alias || "Unknown";
                  const capacity = channel.localBalance + channel.remoteBalance;

                  return (
                    <TableRow key={channel.id} className="channel">
                      <TableCell>
                        {channel.status == "online" ? (
                          <Badge variant="positive">Online</Badge>
                        ) : channel.status == "opening" ? (
                          <Badge variant="outline">Opening</Badge>
                        ) : (
                          <Badge variant="warning">Offline</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <span className="font-medium mr-2">{alias}</span>
                        <Badge variant="outline">
                          {channel.public ? "Public" : "Private"}
                        </Badge>
                      </TableCell>
                      <TableCell title={capacity / 1000 + " sats"}>
                        {formatAmount(capacity)} sats
                      </TableCell>
                      <TableCell
                        title={channel.unspendablePunishmentReserve + " sats"}
                      >
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
                        {formatAmount(
                          channel.unspendablePunishmentReserve * 1000
                        )}{" "}
                        sats
                      </TableCell>
                      <TableCell>
                        <div className="relative">
                          <Progress
                            value={
                              (channel.localSpendableBalance / capacity) * 100
                            }
                            className="h-6 absolute"
                          />
                          <div className="flex flex-row w-full justify-between px-2 text-xs items-center h-6 mix-blend-exclusion text-white">
                            <span
                              title={
                                channel.localSpendableBalance / 1000 + " sats"
                              }
                            >
                              {formatAmount(channel.localSpendableBalance)} sats
                            </span>
                            <span
                              title={channel.remoteBalance / 1000 + " sats"}
                            >
                              {formatAmount(channel.remoteBalance)} sats
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
                          editChannel={editChannel}
                        />
                      </TableCell>
                    </TableRow>
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
    </div>
  );
}
