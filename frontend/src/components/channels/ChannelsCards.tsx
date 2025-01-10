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
import { formatAmount } from "src/lib/utils.ts";
import { Channel, Node } from "src/types";

type ChannelsCardsProps = {
  channels?: Channel[];
  nodes?: Node[];
};

export function ChannelsCards({ channels, nodes }: ChannelsCardsProps) {
  if (!channels?.length) {
    return null;
  }

  return (
    <>
      <Card>
        <CardHeader className="w-full pb-4">Channels</CardHeader>
        <div className="flex flex-col gap-4 slashed-zero p-4">
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
                <>
                  <div className="flex flex-col items-start w-full">
                    <CardTitle className="w-full">
                      <div className="flex items-center justify-between">
                        <div className="flex-1 whitespace-nowrap text-ellipsis font-semibold overflow-hidden">
                          {alias}
                        </div>
                        <ChannelDropdownMenu alias={alias} channel={channel} />
                      </div>
                    </CardTitle>
                    <CardDescription className="w-full flex flex-col gap-4 mt-4">
                      <div className="flex w-full justify-between items-center">
                        <p className="text-muted-foreground font-medium">
                          Status
                        </p>
                        {channel.status == "online" ? (
                          <Badge variant="positive">Online</Badge>
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
                            <TooltipContent className="w-[400px]">
                              The type of lightning channel, By default private
                              channel is recommended. If you a podcaster or
                              musician and expect to receive keysend or
                              Value4Value payments you will need a public
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
                            <TooltipContent className="w-[400px]">
                              Total Spending and Receiving capacity of your
                              lightning channel.
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>

                        <p className="text-foreground">
                          {formatAmount(capacity)} sats
                        </p>
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
                            <TooltipContent className="w-[400px]">
                              Funds each participant sets aside to discourage
                              cheating by ensuring each party has something at
                              stake. This reserve cannot be spent during the
                              channel's lifetime and typically amounts to 1% of
                              the channel capacity.
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
                          {formatAmount(
                            channel.unspendablePunishmentReserve * 1000
                          )}{" "}
                          sats
                        </p>
                      </div>
                    </CardDescription>
                  </div>

                  <CardContent className="p-0">
                    <div className="flex justify-between items-center">
                      <p className="text-muted-foreground font-medium text-sm">
                        Spending
                      </p>
                      <p className="text-muted-foreground font-medium text-sm">
                        Receiving
                      </p>
                    </div>
                    <div className="flex gap-2 items-center mt-2">
                      <div className="flex-1 relative">
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
                          <span title={channel.remoteBalance / 1000 + " sats"}>
                            {formatAmount(channel.remoteBalance)} sats
                          </span>
                        </div>
                      </div>
                      <ChannelWarning channel={channel} />
                    </div>
                  </CardContent>
                  <Separator className="mt-5" />
                </>
              );
            })}
        </div>
      </Card>
    </>
  );
}
