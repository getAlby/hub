import {
  AlertTriangle,
  ExternalLinkIcon,
  HandCoins,
  InfoIcon,
  MoreHorizontal,
  Trash2,
} from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import { Badge } from "src/components/ui/badge.tsx";
import { Button } from "src/components/ui/button.tsx";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu.tsx";
import { Progress } from "src/components/ui/progress.tsx";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip.tsx";
import { formatAmount } from "src/lib/utils.ts";
import { Channel, Node } from "src/types";

type ChannelsCardsProps = {
  channels?: Channel[];
  nodes?: Node[];
  closeChannel(
    channelId: string,
    counterpartyNodeId: string,
    isActive: boolean
  ): void;
  editChannel(channel: Channel): void;
};

export function ChannelsCards({
  channels,
  nodes,
  closeChannel,
  editChannel,
}: ChannelsCardsProps) {
  if (!channels?.length) {
    return null;
  }

  return (
    <>
      <p className="font-semibold text-lg mt-4">Channels</p>
      <div className="flex flex-col gap-2">
        {channels
          .sort((a, b) =>
            a.localBalance + a.remoteBalance > b.localBalance + b.remoteBalance
              ? -1
              : 1
          )
          .map((channel) => {
            const node = nodes?.find(
              (n) => n.public_key === channel.remotePubkey
            );
            const alias = node?.alias || "Unknown";
            const capacity = channel.localBalance + channel.remoteBalance;
            // TODO: remove duplication
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
              <Card>
                <CardHeader className="w-full">
                  <div className="flex flex-col items-start w-full">
                    <CardTitle className="w-full">
                      <div className="flex items-center justify-between">
                        <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden">
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
                        </div>
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
                      </div>
                    </CardTitle>
                    <CardDescription className="w-full flex flex-col gap-4 mt-4">
                      <div className="flex w-full justify-between items-center">
                        <p>Status</p>
                        <div className="flex gap-2">
                          {channelStatus == "online" ? (
                            <Badge variant="positive">Online</Badge>
                          ) : channelStatus == "opening" ? (
                            <Badge variant="outline">Opening</Badge>
                          ) : (
                            <Badge variant="outline">Offline</Badge>
                          )}
                          <Badge variant="outline">
                            {channel.public ? "Public" : "Private"}
                          </Badge>
                        </div>
                      </div>
                      <div className="flex justify-between items-center">
                        <p>Capacity</p>
                        {formatAmount(capacity)} sats
                      </div>
                      <div className="flex justify-between items-center">
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <div className="flex flex-row gap-2 items-center">
                                Reserve
                                <InfoIcon className="h-4 w-4 shrink-0" />
                              </div>
                            </TooltipTrigger>
                            <TooltipContent className="w-[400px]">
                              Funds each participant sets aside to discourage
                              cheating by ensuring each party has something at
                              stake. This reserve cannot be spent during the
                              channel's lifetime and typically amounts to 1% of
                              the channel capacity.
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
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
                      </div>
                    </CardDescription>
                  </div>
                </CardHeader>
                <CardContent className="text-right">
                  <div className="flex gap-2 items-center">
                    <div className="flex-1 relative">
                      <Progress
                        value={(channel.localSpendableBalance / capacity) * 100}
                        className="h-6 absolute"
                      />
                      <div className="flex flex-row w-full justify-between px-2 text-xs items-center h-6 mix-blend-exclusion text-white">
                        <span
                          title={channel.localSpendableBalance / 1000 + " sats"}
                        >
                          {formatAmount(channel.localSpendableBalance)} sats
                        </span>
                        <span title={channel.remoteBalance / 1000 + " sats"}>
                          {formatAmount(channel.remoteBalance)} sats
                        </span>
                      </div>
                    </div>
                    {channelWarning ? (
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <AlertTriangle className="w-4 h-4" />
                          </TooltipTrigger>
                          <TooltipContent className="max-w-[400px]">
                            {channelWarning}
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    ) : null}
                  </div>
                </CardContent>
              </Card>
            );
          })}
      </div>
    </>
  );
}
