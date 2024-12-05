import {
  AlertTriangle,
  Bitcoin,
  ChevronDown,
  CopyIcon,
  Heart,
  Hotel,
  HourglassIcon,
  InfoIcon,
  Settings2,
  Unplug,
  ZapIcon,
} from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader.tsx";
import { ChannelsCards } from "src/components/channels/ChannelsCards.tsx";
import { ChannelsTable } from "src/components/channels/ChannelsTable.tsx";
import EmptyState from "src/components/EmptyState.tsx";
import ExternalLink from "src/components/ExternalLink";
import { TransferFundsButton } from "src/components/TransferFundsButton";
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
import { useIsDesktop } from "src/hooks/useMediaQuery.ts";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { copyToClipboard } from "src/lib/clipboard.ts";
import { cn } from "src/lib/utils.ts";
import { Channel, Node } from "src/types";
import { request } from "src/utils/request";

export default function Channels() {
  useSyncWallet();
  const { data: channels, mutate: reloadChannels } = useChannels();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { data: balances, mutate: reloadBalances } = useBalances();
  const { data: albyBalance, mutate: reloadAlbyBalance } = useAlbyBalance();
  const navigate = useNavigate();
  const [nodes, setNodes] = React.useState<Node[]>([]);

  const { toast } = useToast();
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

  const showHostedBalance =
    albyBalance && albyBalance.sats > ALBY_HIDE_HOSTED_BALANCE_LIMIT;

  return (
    <>
      <AppHeader
        title="Node"
        description="Manage your lightning node"
        contentRight={
          <div className="flex gap-3 items-center justify-center">
            <DropdownMenu modal={false}>
              <DropdownMenuTrigger asChild>
                {isDesktop ? (
                  <Button
                    className="inline-flex"
                    variant="outline"
                    size="default"
                  >
                    Advanced
                    <ChevronDown />
                  </Button>
                ) : (
                  <Button variant="outline" size="icon">
                    <Settings2 className="w-4 h-4" />
                  </Button>
                )}
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
                        copyToClipboard(nodeConnectionInfo.pubkey, toast);
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
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuLabel>On-Chain</DropdownMenuLabel>
                  <DropdownMenuItem>
                    <Link
                      to="/channels/onchain/deposit-bitcoin"
                      className="w-full"
                    >
                      Deposit Bitcoin
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <Link to="onchain/buy-bitcoin" className="w-full">
                      Buy Bitcoin
                    </Link>
                  </DropdownMenuItem>
                  {(balances?.onchain.spendable || 0) > ONCHAIN_DUST_SATS && (
                    <DropdownMenuItem
                      onClick={() => navigate("/wallet/withdraw")}
                      className="w-full cursor-pointer"
                    >
                      Withdraw On-Chain Balance
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
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
            <Link to="/channels/incoming">
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
        </>
      )}

      <div
        className={cn("flex flex-col sm:flex-row flex-wrap gap-3 slashed-zero")}
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
                {new Intl.NumberFormat().format(albyBalance.sats)} sats
              </div>
            </CardContent>
            <CardFooter className="flex justify-end space-x-1">
              <TransferFundsButton
                variant="outline"
                channels={channels}
                albyBalance={albyBalance}
                onTransferComplete={() =>
                  Promise.all([
                    reloadAlbyBalance(),
                    reloadBalances(),
                    reloadChannels(),
                  ])
                }
              >
                Transfer
              </TransferFundsButton>
            </CardFooter>
          </Card>
        )}

        <Card className="flex flex-1 sm:flex-[2] flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="font-semibold">Lightning</CardTitle>
            <ZapIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>

          <CardContent className="flex flex-col sm:flex-row pl-0 flex-wrap">
            <div className="flex flex-col flex-1">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 pr-0">
                <CardTitle className="text-sm font-medium">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <div className="flex flex-row gap-2 items-center justify-start text-sm">
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
              </CardHeader>
              <CardContent className="flex-grow pb-0">
                {!balances && (
                  <div>
                    <div className="animate-pulse d-inline ">
                      <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                    </div>
                  </div>
                )}
                {balances && (
                  <div className="text-2xl font-bold balance sensitive">
                    {new Intl.NumberFormat().format(
                      Math.floor(balances.lightning.totalSpendable / 1000)
                    )}{" "}
                    sats
                  </div>
                )}
              </CardContent>
            </div>
            <div className="flex flex-col flex-1">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 pr-0">
                <CardTitle className="text-sm font-medium">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <div className="flex flex-row gap-2 items-center justify-start text-sm">
                          Receiving Capacity
                          <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                        </div>
                      </TooltipTrigger>
                      <TooltipContent className="w-[300px]">
                        Your receiving capacity is the funds owned by your
                        channel partner, which will be moved to your side when
                        you receive lightning payments.
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </CardTitle>
              </CardHeader>
              <CardContent className="flex-grow pb-0">
                <div className="text-2xl font-bold balance sensitive">
                  {balances &&
                    new Intl.NumberFormat().format(
                      Math.floor(balances.lightning.totalReceivable / 1000)
                    )}{" "}
                  sats
                </div>
              </CardContent>
            </div>
          </CardContent>
        </Card>
        <Card className="flex flex-1 flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="font-semibold">On-Chain</CardTitle>

            <Bitcoin className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex-grow pb-0">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 pl-0">
              <CardTitle className="text-sm font-medium">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <div className="flex flex-row gap-2 items-center text-sm">
                        Balance
                        <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                      </div>
                    </TooltipTrigger>
                    <TooltipContent className="w-[300px]">
                      Your on-chain balance can be used to open new outgoing
                      lightning channels. When channels are closed, funds on
                      your side of your channel will be returned to your savings
                      balance.
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </CardTitle>
            </CardHeader>
            {!balances && (
              <div>
                <div className="animate-pulse d-inline ">
                  <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                </div>
              </div>
            )}
            <div className="text-2xl font-bold balance sensitive">
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
        </Card>
      </div>

      {balances && balances.onchain.pendingBalancesFromChannelClosures > 0 && (
        <Alert>
          <HourglassIcon className="h-4 w-4" />
          <AlertTitle>Pending Closed Channels</AlertTitle>
          <AlertDescription>
            You have{" "}
            {new Intl.NumberFormat().format(
              balances.onchain.pendingBalancesFromChannelClosures
            )}{" "}
            sats pending from one or more closed channels. Once spendable again
            these will become available in your on-chain balance. Funds from
            channels that were force closed may take up to 2 weeks to become
            available.{" "}
            <ExternalLink
              to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/faq-alby-hub/why-was-my-lightning-channel-closed-and-what-to-do-next"
              className="underline"
            >
              Learn more
            </ExternalLink>
          </AlertDescription>
        </Alert>
      )}

      {channels && channels.length === 0 && (
        <EmptyState
          icon={Unplug}
          title="No Channels Available"
          description="Connect to the Lightning Network by establishing your first channel and start transacting."
          buttonText="Open Channel"
          buttonLink="/channels/incoming"
        />
      )}

      {isDesktop ? (
        <ChannelsTable channels={channels} nodes={nodes} />
      ) : (
        <ChannelsCards channels={channels} nodes={nodes} />
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
