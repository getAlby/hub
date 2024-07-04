import {
  AlertTriangle,
  ArrowDown,
  ArrowUp,
  Bitcoin,
  ChevronDown,
  CopyIcon,
  ExternalLinkIcon,
  HandCoins,
  Hotel,
  InfoIcon,
  MoreHorizontal,
  Trash2,
  Unplug,
} from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader.tsx";
import EmptyState from "src/components/EmptyState.tsx";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading.tsx";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Badge } from "src/components/ui/badge.tsx";
import { Button } from "src/components/ui/button.tsx";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card.tsx";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu.tsx";
import { LoadingButton } from "src/components/ui/loading-button.tsx";
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
import { useToast } from "src/components/ui/use-toast.ts";
import {
  ALBY_HIDE_HOSTED_BALANCE_BELOW as ALBY_HIDE_HOSTED_BALANCE_LIMIT,
  ONCHAIN_DUST_SATS,
} from "src/constants.ts";
import { useAlbyBalance } from "src/hooks/useAlbyBalance.ts";
import { useBalances } from "src/hooks/useBalances.ts";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo.ts";
import { useRedeemOnchainFunds } from "src/hooks/useRedeemOnchainFunds.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { copyToClipboard } from "src/lib/clipboard.ts";
import { cn, formatAmount } from "src/lib/utils.ts";
import {
  Channel,
  CloseChannelResponse,
  Node,
  UpdateChannelRequest,
} from "src/types";
import { request } from "src/utils/request";
import { useCSRF } from "../../hooks/useCSRF.ts";

export default function Channels() {
  useSyncWallet();
  const { data: channels, mutate: reloadChannels } = useChannels();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { data: balances } = useBalances();
  const { data: albyBalance, mutate: reloadAlbyBalance } = useAlbyBalance();
  const [nodes, setNodes] = React.useState<Node[]>([]);
  const { mutate: reloadInfo } = useInfo();
  const { data: csrf } = useCSRF();
  const redeemOnchainFunds = useRedeemOnchainFunds();
  const { toast } = useToast();
  const [drainingAlbySharedFunds, setDrainingAlbySharedFunds] =
    React.useState(false);

  // TODO: move to NWC backend
  const loadNodeStats = React.useCallback(async () => {
    if (!channels) {
      return [];
    }
    const nodes = await Promise.all(
      channels?.map(async (channel): Promise<Node | undefined> => {
        try {
          const response = await request<Node>(
            `/api/mempool?endpoint=/v1/lightning/nodes/${channel.remotePubkey}`
          );
          return response;
        } catch (error) {
          console.error(error);
          return undefined;
        }
      })
    );
    setNodes(nodes.filter((node) => !!node) as Node[]);
  }, [channels]);

  React.useEffect(() => {
    loadNodeStats();
  }, [loadNodeStats]);

  async function closeChannel(
    channelId: string,
    nodeId: string,
    isActive: boolean
  ) {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      if (!isActive) {
        if (
          !confirm(
            `This channel is inactive. Some channels require up to 6 onchain confirmations before they are usable. If you really want to continue, click OK.`
          )
        ) {
          return;
        }
      }

      if (
        !confirm(
          `Are you sure you want to close the channel with ${
            nodes.find((node) => node.public_key === nodeId)?.alias ||
            "Unknown Node"
          }?\n\nNode ID: ${nodeId}\n\nChannel ID: ${channelId}`
        )
      ) {
        return;
      }

      const closeType = prompt(
        "Select way to close the channel. Type 'force close' if you want to force close the channel. Note: your channel balance will be locked for up to two weeks if you force close.",
        "normal close"
      );
      if (!closeType) {
        console.error("Cancelled close channel");
        return;
      }

      console.info(`🎬 Closing channel with ${nodeId}`);

      const closeChannelResponse = await request<CloseChannelResponse>(
        `/api/peers/${nodeId}/channels/${channelId}?force=${
          closeType === "force close"
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
        (c) => c.id === channelId && c.remotePubkey === nodeId
      );
      console.info("Closed channel", closedChannel);
      if (closedChannel) {
        prompt(
          "Closed channel. Copy channel funding TX to view on mempool",
          closedChannel.fundingTxId
        );
      }
      await reloadChannels();
      toast({ title: "Sucessfully closed channel" });
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        description: "Something went wrong: " + error,
      });
    }
  }

  async function editChannel(channel: Channel) {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }

      const forwardingFeeBaseSats = prompt(
        "Enter base forwarding fee in sats",
        Math.floor(channel.forwardingFeeBaseMsat / 1000).toString()
      );

      if (!forwardingFeeBaseSats) {
        return;
      }

      const forwardingFeeBaseMsat = +forwardingFeeBaseSats * 1000;

      console.info(
        `🎬 Updating channel ${channel.id} with ${channel.remotePubkey}`
      );

      await request(
        `/api/peers/${channel.remotePubkey}/channels/${channel.id}`,
        {
          method: "PATCH",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            forwardingFeeBaseMsat: forwardingFeeBaseMsat,
          } as UpdateChannelRequest),
        }
      );
      await reloadChannels();
      toast({ title: "Sucessfully updated channel" });
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        description: "Something went wrong: " + error,
      });
    }
  }

  async function resetRouter() {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }

      const key = prompt(
        "Enter key to reset (choose one of ALL, LatestRgsSyncTimestamp, Scorer, NetworkGraph). After resetting, you'll need to re-enter your unlock password.",
        "ALL"
      );
      if (!key) {
        console.error("Cancelled reset");
        return;
      }

      await request("/api/reset-router", {
        method: "POST",
        body: JSON.stringify({ key }),
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
      });
      await reloadInfo();
      toast({ description: "🎉 Router reset" });
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        description: "Something went wrong: " + error,
      });
    }
  }

  const showHostedBalance =
    albyBalance && albyBalance.sats > ALBY_HIDE_HOSTED_BALANCE_LIMIT;

  return (
    <>
      <AppHeader
        title="Liquidity"
        description="Manage your lightning node liquidity"
        contentRight={
          <>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="default">
                  Advanced
                  <ChevronDown />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                <DropdownMenuGroup>
                  <DropdownMenuItem>
                    <div
                      className="flex flex-row gap-4 items-center w-full cursor-pointer"
                      onClick={() => {
                        if (!nodeConnectionInfo) {
                          return;
                        }
                        copyToClipboard(nodeConnectionInfo.pubkey);
                      }}
                    >
                      <div>Node</div>
                      <div className="overflow-hidden text-ellipsis flex-1">
                        {nodeConnectionInfo?.pubkey || "Loading..."}
                      </div>
                      {nodeConnectionInfo && (
                        <CopyIcon className="shrink-0 w-4 h-4" />
                      )}
                    </div>
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                <DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem>
                    <Link
                      to="/channels/onchain/deposit-bitcoin"
                      className="w-full"
                    >
                      Deposit Bitcoin
                    </Link>
                  </DropdownMenuItem>
                  {(balances?.onchain.spendable || 0) > ONCHAIN_DUST_SATS && (
                    <DropdownMenuItem
                      onClick={redeemOnchainFunds.redeemFunds}
                      disabled={redeemOnchainFunds.isLoading}
                      className="w-full cursor-pointer"
                    >
                      Redeem Onchain Funds
                      {redeemOnchainFunds.isLoading && <Loading />}
                    </DropdownMenuItem>
                  )}
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuLabel>Management</DropdownMenuLabel>
                  <DropdownMenuItem>
                    <Link className="w-full" to="/peers">
                      Connected Peers
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <Link className="w-full" to="/wallet/sign-message">
                      Sign Message
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="w-full cursor-pointer"
                    onClick={resetRouter}
                  >
                    Clear Routing Data
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
            {/* <Link to="/channels/new">
              <Button>Open Channel</Button>
            </Link> */}
          </>
        }
      ></AppHeader>

      {!!channels?.length && (
        <>
          {/* If all channels have less than 20% incoming capacity, show a warning */}
          {channels?.every(
            (channel) =>
              channel.remoteBalance <
              (channel.localBalance + channel.remoteBalance) * 0.2
          ) && (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Low receiving capacity</AlertTitle>
              <AlertDescription>
                You likely won't be able to receive payments until you{" "}
                <Link className="underline" to="/channels/incoming">
                  increase your receiving capacity.
                </Link>
              </AlertDescription>
            </Alert>
          )}

          {/* If all channels have less or equal balance than their reserve, show a warning */}
          {channels?.every(
            (channel) =>
              channel.localBalance <=
              channel.unspendablePunishmentReserve * 1000
          ) && (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Channel reserves unmet</AlertTitle>
              <AlertDescription>
                You won't be able to make payments until you{" "}
                <Link className="underline" to="/channels/outgoing">
                  increase your spending balance.
                </Link>
              </AlertDescription>
            </Alert>
          )}
        </>
      )}

      <div
        className={cn(
          "grid grid-cols-1 gap-3",
          showHostedBalance ? "xl:grid-cols-4" : "lg:grid-cols-3"
        )}
      >
        {showHostedBalance && (
          <Card className="flex flex-col">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Alby Hosted Balance
              </CardTitle>
              <Hotel className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent className="flex-grow">
              <div className="text-2xl font-bold">
                {new Intl.NumberFormat().format(albyBalance?.sats)} sats
              </div>
            </CardContent>
            <CardFooter className="flex justify-end space-x-1">
              <LoadingButton
                loading={drainingAlbySharedFunds}
                onClick={async () => {
                  if (
                    !channels?.some(
                      (channel) =>
                        channel.remoteBalance / 1000 > albyBalance.sats
                    )
                  ) {
                    toast({
                      title: "Please increase your receiving capacity first",
                    });
                    return;
                  }

                  setDrainingAlbySharedFunds(true);
                  try {
                    if (!csrf) {
                      throw new Error("csrf not loaded");
                    }

                    await request("/api/alby/drain", {
                      method: "POST",
                      headers: {
                        "X-CSRF-Token": csrf,
                        "Content-Type": "application/json",
                      },
                    });
                    await reloadAlbyBalance();
                    toast({
                      description:
                        "🎉 Funds from Alby shared wallet transferred to your Alby Hub!",
                    });
                  } catch (error) {
                    console.error(error);
                    toast({
                      variant: "destructive",
                      description: "Something went wrong: " + error,
                    });
                  }
                  setDrainingAlbySharedFunds(false);
                }}
                variant="outline"
              >
                Transfer
              </LoadingButton>
            </CardFooter>
          </Card>
        )}
        <Card className="flex flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Savings Balance
            </CardTitle>
            <Bitcoin className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex-grow">
            {!balances && (
              <div>
                <div className="animate-pulse d-inline ">
                  <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                </div>
              </div>
            )}
            <div className="text-2xl font-bold">
              {balances && (
                <>
                  {new Intl.NumberFormat().format(balances.onchain.spendable)}{" "}
                  sats
                  {balances &&
                    balances.onchain.spendable !== balances.onchain.total && (
                      <p className="text-xs text-muted-foreground animate-pulse">
                        +
                        {new Intl.NumberFormat().format(
                          balances.onchain.total - balances.onchain.spendable
                        )}{" "}
                        sats incoming
                      </p>
                    )}
                </>
              )}
            </div>
          </CardContent>
          <CardFooter className="flex justify-end space-x-1">
            <Link to="onchain/buy-bitcoin">
              <Button variant="outline">Buy</Button>
            </Link>
            <Link to="onchain/deposit-bitcoin">
              <Button variant="outline">Deposit</Button>
            </Link>
          </CardFooter>
        </Card>
        <Card className="flex flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Spending Balance
            </CardTitle>
            <ArrowUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex-grow">
            {!balances && (
              <div>
                <div className="animate-pulse d-inline ">
                  <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                </div>
              </div>
            )}
            {balances && (
              <div className="text-2xl font-bold">
                {new Intl.NumberFormat(undefined, {}).format(
                  Math.floor(balances.lightning.totalSpendable / 1000)
                )}{" "}
                sats
              </div>
            )}
          </CardContent>
          <CardFooter className="flex justify-end">
            <Link to="/channels/outgoing">
              <Button variant="outline">Top Up</Button>
            </Link>
          </CardFooter>
        </Card>
        <Card className="flex flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Receiving Capacity
            </CardTitle>
            <ArrowDown className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex-grow">
            <div className="text-2xl font-bold">
              {balances &&
                new Intl.NumberFormat().format(
                  Math.floor(balances.lightning.totalReceivable / 1000)
                )}{" "}
              sats
            </div>
          </CardContent>
          <CardFooter className="flex justify-end">
            <Link to="/channels/incoming">
              <Button variant="outline">Increase</Button>
            </Link>
          </CardFooter>
        </Card>
      </div>

      {channels && channels.length === 0 && (
        <EmptyState
          icon={Unplug}
          title="No Channels Available"
          description="Connect to the Lightning Network by establishing your first channel and start transacting."
          buttonText="Open Channel"
          buttonLink="/channels/outgoing"
        />
      )}

      {!channels ||
        (channels.length > 0 && (
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
                        Funds each participant sets aside to discourage cheating
                        by ensuring each party has something at stake. This
                        reserve cannot be spent during the channel's lifetime
                        and typically amounts to 1% of the channel capacity.
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </TableHead>
                <TableHead className="w-[300px]">
                  <div className="flex flex-row justify-between items-center">
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
                      const node = nodes.find(
                        (n) => n.public_key === channel.remotePubkey
                      );
                      const alias = node?.alias || "Unknown";
                      const capacity =
                        channel.localBalance + channel.remoteBalance;

                      let channelWarning = "";
                      if (channel.localSpendableBalance < capacity * 0.1) {
                        channelWarning =
                          "Spending balance low. You may have trouble sending payments through this channel.";
                      }
                      if (channel.localSpendableBalance > capacity * 0.9) {
                        channelWarning =
                          "Receiving capacity low. You may have trouble receiving payments through this channel.";
                      }

                      return (
                        <TableRow key={channel.id}>
                          <TableCell>
                            {channel.active ? (
                              <Badge variant="positive">Online</Badge>
                            ) : (
                              <Badge variant="outline">Offline</Badge>
                            )}{" "}
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
                            title={
                              channel.unspendablePunishmentReserve + " sats"
                            }
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
                                  (channel.localSpendableBalance / capacity) *
                                  100
                                }
                                className="h-6 absolute"
                              />
                              <div className="flex flex-row w-full justify-between px-2 text-xs items-center h-6 mix-blend-exclusion text-white">
                                <span
                                  title={
                                    channel.localSpendableBalance / 1000 +
                                    " sats"
                                  }
                                >
                                  {formatAmount(channel.localSpendableBalance)}{" "}
                                  sats
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
                                  <TooltipContent className="w-[400px]">
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
        ))}
    </>
  );
}
