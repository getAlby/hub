import {
  AlertTriangle,
  ExternalLinkIcon,
  HandCoins,
  InfoIcon,
  MoreHorizontal,
  Trash2,
} from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading.tsx";
import { Badge } from "src/components/ui/badge.tsx";
import { Button } from "src/components/ui/button.tsx";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu.tsx";
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

type ChannelsTableProps = {
  channels?: Channel[];
  nodes?: Node[];
  closeChannel(
    channelId: string,
    counterpartyNodeId: string,
    isActive: boolean
  ): void;
  editChannel(channel: Channel): void;
};

export function ChannelsTable({
  channels,
  nodes,
  closeChannel,
  editChannel,
}: ChannelsTableProps) {
  if (channels && !channels.length) {
    return null;
  }

  return (
    <div className="border border-red-500 max-w-full overflow-y-auto">
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

                  let channelWarning = "";
                  if (channel.error) {
                    channelWarning = channel.error;
                  } else {
                    if (channel.localSpendableBalance < capacity * 0.1) {
                      channelWarning =
                        "Spending balance low. You may have trouble sending payments through this channel.";
                    }
                    if (channel.localSpendableBalance > capacity * 0.9) {
                      channelWarning =
                        "Receiving capacity low. You may have trouble receiving payments through this channel.";
                    }
                  }

                  const channelStatus = channel.active
                    ? "online"
                    : channel.confirmationsRequired !== undefined &&
                        channel.confirmations !== undefined &&
                        channel.confirmationsRequired > channel.confirmations
                      ? "opening"
                      : "offline";
                  if (channelStatus === "opening") {
                    channelWarning = `Channel is currently being opened (${channel.confirmations} of ${channel.confirmationsRequired} confirmations). Once the required confirmation are reached, you will be able to send and receive on this channel.`;
                  }
                  if (channelStatus === "offline") {
                    channelWarning =
                      "This channel is currently offline and cannot be used to send or receive payments. Please contact Alby Support for more information.";
                  }

                  return (
                    <TableRow key={channel.id} className="channel">
                      <TableCell>
                        {channelStatus == "online" ? (
                          <Badge variant="positive">Online</Badge>
                        ) : channelStatus == "opening" ? (
                          <Badge variant="outline">Opening</Badge>
                        ) : (
                          <Badge variant="outline">Offline</Badge>
                        )}
                      </TableCell>
                      <TableCell className="flex flex-row items-center">
                        <a
                          title={channel.remotePubkey}
                          href={`https://amboss.space/node/${channel.remotePubkey}`}
                          target="_blank"
                          rel="noopener noreferer"
                        >
                          <Button variant="link" className="p-0 mr-2">
                            {alias}
                          </Button>
                        </a>
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
                        {channelWarning ? (
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger>
                                <AlertTriangle className="w-4 h-4 mt-1" />
                              </TooltipTrigger>
                              <TooltipContent className="max-w-[400px]">
                                {channelWarning}
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        ) : null}
                      </TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button size="icon" variant="ghost">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem className="flex flex-row items-center gap-2 cursor-pointer">
                              <ExternalLink
                                to={`https://mempool.space/tx/${channel.fundingTxId}`}
                                className="w-full flex flex-row items-center gap-2"
                              >
                                <ExternalLinkIcon className="w-4 h-4" />
                                <p>View Funding Transaction</p>
                              </ExternalLink>
                            </DropdownMenuItem>
                            {channel.public && (
                              <DropdownMenuItem
                                className="flex flex-row items-center gap-2 cursor-pointer"
                                onClick={() => editChannel(channel)}
                              >
                                <HandCoins className="h-4 w-4" />
                                Set Routing Fee
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuItem
                              className="flex flex-row items-center gap-2 cursor-pointer"
                              onClick={() =>
                                closeChannel(
                                  channel.id,
                                  channel.remotePubkey,
                                  channel.active
                                )
                              }
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                              Close Channel
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
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
