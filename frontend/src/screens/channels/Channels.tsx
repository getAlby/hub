import {
  AlertTriangle,
  ArrowDown,
  ArrowUp,
  Bitcoin,
  ChevronDown,
  CopyIcon,
  Heart,
  Hotel,
  InfoIcon,
  Unplug,
} from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader.tsx";
import { ChannelsCards } from "src/components/channels/ChannelsCards.tsx";
import { ChannelsTable } from "src/components/channels/ChannelsTable.tsx";
import EmptyState from "src/components/EmptyState.tsx";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading.tsx";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
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
import { CircleProgress } from "src/components/ui/progress.tsx";
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
import { useIsDesktop } from "src/hooks/useMediaQuery.ts";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo.ts";
import { useRedeemOnchainFunds } from "src/hooks/useRedeemOnchainFunds.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { copyToClipboard } from "src/lib/clipboard.ts";
import { cn } from "src/lib/utils.ts";
import { Channel, Node, UpdateChannelRequest } from "src/types";
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
  const isDesktop = useIsDesktop();

  const nodeHealth = channels ? getNodeHealth(channels) : 0;

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
        title="Node"
        description="Manage your lightning node"
        contentRight={
          <div className="flex gap-3 items-center justify-center">
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
                      Withdraw Savings Balance
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
            <Link to="/channels/outgoing">
              <Button>Open Channel</Button>
            </Link>
            <ExternalLink to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/liquidity/node-health">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <CircleProgress
                      value={nodeHealth}
                      className="w-9 h-9 relative"
                    >
                      {nodeHealth === 100 && (
                        <div className="absolute w-full h-full opacity-20">
                          <div className="absolute w-full h-full bg-primary animate-pulse" />
                        </div>
                      )}
                      <Heart
                        className="w-4 h-4"
                        stroke={"hsl(var(--primary))"}
                        strokeWidth={3}
                        fill={
                          nodeHealth === 100
                            ? "hsl(var(--primary))"
                            : "transparent"
                        }
                      />
                    </CircleProgress>
                  </TooltipTrigger>
                  <TooltipContent>Node health: {nodeHealth}%</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </ExternalLink>
          </div>
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
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <div className="flex flex-row gap-2 items-center">
                      Savings Balance
                      <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                    </div>
                  </TooltipTrigger>
                  <TooltipContent className="w-[300px]">
                    Your savings balance (on-chain balance) can be used to open
                    new lightning channels. When channels are closed, funds from
                    your channel will be returned to your savings balance.
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
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
            <div className="text-2xl font-bold balance sensitive ph-no-capture">
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
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <div className="flex flex-row gap-2 items-center">
                      Spending Balance
                      <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                    </div>
                  </TooltipTrigger>
                  <TooltipContent className="w-[300px]">
                    Your spending balance is the funds on your side of your
                    channels, which you can use to make lightning payments.
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
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
              <div className="text-2xl font-bold balance sensitive ph-no-capture">
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
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <div className="flex flex-row gap-2 items-center">
                      Receiving Capacity
                      <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                    </div>
                  </TooltipTrigger>
                  <TooltipContent className="w-[300px]">
                    Your receiving capacity is the funds owned by your channel
                    partner, which will be moved to your side when you receive
                    lightning payments.
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </CardTitle>
            <ArrowDown className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex-grow">
            <div className="text-2xl font-bold balance sensitive ph-no-capture">
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

      {isDesktop ? (
        <ChannelsTable
          channels={channels}
          nodes={nodes}
          editChannel={editChannel}
        />
      ) : (
        <ChannelsCards
          channels={channels}
          nodes={nodes}
          editChannel={editChannel}
        />
      )}
    </>
  );
}

function getNodeHealth(channels: Channel[]) {
  const totalChannelCapacitySats = channels
    .map((channel) => (channel.localBalance + channel.remoteBalance) / 1000)
    .reduce((a, b) => a + b, 0);
  const averageChannelBalance =
    channels
      .map((channel) => {
        const totalBalance = channel.localBalance + channel.remoteBalance;
        const expectedBalance = totalBalance / 2;
        const actualBalance =
          Math.min(channel.localBalance, channel.remoteBalance) /
          expectedBalance;
        return actualBalance;
      })
      .reduce((a, b) => a + b, 0) / (channels.length || 1);

  const numUniqueChannelPartners = new Set(
    channels.map((channel) => channel.remotePubkey)
  ).size;

  const nodeHealth = Math.ceil(
    numUniqueChannelPartners *
      (100 / 2) * // 2 or more channels is great
      (Math.min(totalChannelCapacitySats, 1_000_000) / 1_000_000) * // 1 million sats or more is great
      (0.9 + averageChannelBalance * 0.1) // +10% for perfectly balanced channels
  );

  if (nodeHealth > 95) {
    // prevent OCD
    return 100;
  }

  return nodeHealth;
}
