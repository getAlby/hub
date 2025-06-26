import dayjs from "dayjs";
import {
  AlertTriangleIcon,
  ArrowRightIcon,
  CopyIcon,
  ExternalLinkIcon,
  HeartIcon,
  HourglassIcon,
  InfoIcon,
  LinkIcon,
  Settings2Icon,
  SparklesIcon,
  UnplugIcon,
  ZapIcon,
} from "lucide-react";
import React from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import AppHeader from "src/components/AppHeader.tsx";
import { ChannelsCards } from "src/components/channels/ChannelsCards.tsx";
import { ChannelsTable } from "src/components/channels/ChannelsTable.tsx";
import { HealthCheckAlert } from "src/components/channels/HealthcheckAlert";
import { OnchainTransactionsTable } from "src/components/channels/OnchainTransactionsTable.tsx";
import { SwapDialogs } from "src/components/channels/SwapDialogs";
import EmptyState from "src/components/EmptyState.tsx";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import ResponsiveButton from "src/components/ResponsiveButton";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Button } from "src/components/ui/button.tsx";
import {
  Card,
  CardContent,
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
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { ONCHAIN_DUST_SATS } from "src/constants.ts";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances.ts";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { copyToClipboard } from "src/lib/clipboard.ts";
import { cn } from "src/lib/utils.ts";
import {
  Channel,
  LongUnconfirmedZeroConfChannel,
  MempoolNode,
  MempoolTransaction,
} from "src/types";
import { request } from "src/utils/request";

export default function Channels() {
  useSyncWallet();
  const { data: info } = useInfo();
  const { data: albyMe } = useAlbyMe();
  const { data: channels } = useChannels();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const navigate = useNavigate();
  const [nodes, setNodes] = React.useState<MempoolNode[]>([]);
  const [longUnconfirmedZeroConfChannels, setLongUnconfirmedZeroConfChannels] =
    React.useState<LongUnconfirmedZeroConfChannel[]>([]);
  const [swapOutDialogOpen, setSwapOutDialogOpen] = React.useState(false);
  const [swapInDialogOpen, setSwapInDialogOpen] = React.useState(false);
  const [searchParams, setSearchParams] = useSearchParams();

  React.useEffect(() => {
    if (balances && channels && searchParams.has("swap", "true")) {
      setSearchParams({});
      if (
        balances.lightning.totalSpendable > balances.lightning.totalReceivable
      ) {
        setSwapOutDialogOpen(true);
      } else {
        setSwapInDialogOpen(true);
      }
    }
  }, [balances, channels, searchParams, setSearchParams]);

  const { toast } = useToast();

  const nodeHealth = channels ? getNodeHealth(channels) : 0;

  // TODO: move to NWC backend
  const loadNodeStats = React.useCallback(async () => {
    if (!channels) {
      return [];
    }
    const nodes = await Promise.all(
      channels?.map(async (channel): Promise<MempoolNode | undefined> => {
        try {
          const response = await request<MempoolNode>(
            `/api/mempool?endpoint=/v1/lightning/nodes/${channel.remotePubkey}`
          );
          return response;
        } catch (error) {
          console.error(error);
          return undefined;
        }
      })
    );
    setNodes(nodes.filter((node) => !!node) as MempoolNode[]);
  }, [channels]);

  const findUnconfirmedChannels = React.useCallback(async () => {
    if (!channels) {
      return [];
    }

    const _longUnconfirmedZeroConfChannels: LongUnconfirmedZeroConfChannel[] =
      [];
    for (const channel of channels) {
      // only check for unconfirmed 0-conf active channels
      if (
        channel.status !== "online" ||
        channel.confirmationsRequired !== 0 ||
        !!channel.confirmations
      ) {
        continue;
      }
      try {
        const mempoolTransactionResponse = await request<MempoolTransaction>(
          `/api/mempool?endpoint=/tx/${channel.fundingTxId}`
        );
        if (!mempoolTransactionResponse) {
          throw new Error("No response");
        }
        if (mempoolTransactionResponse.status.confirmed) {
          continue;
        }
        // see if the transaction is in the mempool, and has been for how long
        const unconfirmedTransactionTimeResponse = await request<number[]>(
          `/api/mempool?endpoint=/v1/transaction-times?txId[]=${channel.fundingTxId}`
        );
        if (!unconfirmedTransactionTimeResponse?.length) {
          throw new Error("No response");
        }

        const timestamp = unconfirmedTransactionTimeResponse[0];
        const unconfirmedHours = dayjs().diff(timestamp * 1000, "hours");
        if (unconfirmedHours === 0) {
          // channel has been unconfirmed for less than an hour
          continue;
        }

        _longUnconfirmedZeroConfChannels.push({
          id: channel.id,
          message: "Unconfirmed for " + unconfirmedHours + " hours",
        });
      } catch (error) {
        _longUnconfirmedZeroConfChannels.push({
          id: channel.id,
          message: "Channel transaction not in the mempool yet",
        });
      }
    }
    setLongUnconfirmedZeroConfChannels(_longUnconfirmedZeroConfChannels);
  }, [channels]);

  React.useEffect(() => {
    loadNodeStats();
  }, [loadNodeStats]);

  React.useEffect(() => {
    findUnconfirmedChannels();
  }, [findUnconfirmedChannels]);

  return (
    <>
      <AppHeader
        title="Node"
        contentRight={
          hasChannelManagement && (
            <div className="flex gap-3 items-center justify-center">
              <SwapDialogs
                setSwapOutDialogOpen={setSwapOutDialogOpen}
                swapOutDialogOpen={swapOutDialogOpen}
                setSwapInDialogOpen={setSwapInDialogOpen}
                swapInDialogOpen={swapInDialogOpen}
              />
              <DropdownMenu modal={false}>
                <DropdownMenuTrigger>
                  <ResponsiveButton
                    icon={Settings2Icon}
                    text="Advanced"
                    variant="outline"
                  />
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
                    <DropdownMenuLabel>Swaps</DropdownMenuLabel>
                    <DropdownMenuItem
                      onClick={() => setSwapInDialogOpen(true)}
                      className="cursor-pointer"
                    >
                      <div className="mr-2 text-muted-foreground flex flex-row items-center">
                        <LinkIcon className="w-4 h-4" />
                        <ArrowRightIcon className="w-4 h-4" />
                        <ZapIcon className="w-4 h-4" />
                      </div>
                      Swap in
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => setSwapOutDialogOpen(true)}
                      className="cursor-pointer"
                    >
                      <div className="mr-2 text-muted-foreground flex flex-row items-center">
                        <ZapIcon className="w-4 h-4" />
                        <ArrowRightIcon className="w-4 h-4" />
                        <LinkIcon className="w-4 h-4" />
                      </div>
                      Swap out
                    </DropdownMenuItem>
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
                    {info?.backendType === "LDK" &&
                      (!albyMe?.subscription.plan_code ? (
                        <UpgradeDialog>
                          <div className="cursor-pointer">
                            <DropdownMenuItem className="w-full pointer-events-none">
                              <Link
                                className="w-full flex items-center"
                                to="/wallet/node-alias"
                              >
                                <SparklesIcon className="w-4 h-4 mr-2" /> Set
                                Node Alias
                              </Link>
                            </DropdownMenuItem>
                          </div>
                        </UpgradeDialog>
                      ) : (
                        <DropdownMenuItem className="w-full">
                          <Link className="w-full" to="/wallet/node-alias">
                            Set Node Alias
                          </Link>
                        </DropdownMenuItem>
                      ))}
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>

              <Link to="/channels/incoming">
                <Button>Open Channel</Button>
              </Link>
              <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/node/node-health">
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
                        <HeartIcon
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
          )
        }
      ></AppHeader>

      <HealthCheckAlert />

      {hasChannelManagement && (
        <>
          {!!channels?.length && (
            <>
              {/* If all channels have less than 20% incoming capacity, show a warning */}
              {channels?.every(
                (channel) =>
                  channel.remoteBalance <
                  (channel.localBalance + channel.remoteBalance) * 0.2
              ) && (
                <Alert>
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>Low receiving limit</AlertTitle>
                  <AlertDescription>
                    You likely won't be able to receive payments until you{" "}
                    <Link
                      className="underline"
                      to="#"
                      onClick={() => setSwapOutDialogOpen(true)}
                    >
                      swap out funds
                    </Link>{" "}
                    or{" "}
                    <Link className="underline" to="/channels/incoming">
                      increase your receiving limits
                    </Link>
                    .
                  </AlertDescription>
                </Alert>
              )}
            </>
          )}

          <div
            className={cn(
              "flex flex-col sm:flex-row flex-wrap gap-3 slashed-zero"
            )}
          >
            <Card className="flex flex-1 sm:flex-[2] flex-col">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="font-semibold text-2xl">
                  Lightning
                </CardTitle>
                <ZapIcon className="h-6 w-6 text-muted-foreground" />
              </CardHeader>

              <CardContent className="flex flex-col sm:flex-row pl-0 flex-wrap">
                <div className="flex flex-col flex-1">
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1 pr-0">
                    <CardTitle className="text-sm font-medium">
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <div className="flex flex-row gap-1 items-center justify-start text-sm font-medium">
                              Spending Balance
                              <InfoIcon className="h-3 w-3 shrink-0 text-muted-foreground" />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent className="w-[300px]">
                            Your spending balance is the funds on your side of
                            your channels, which you can use to make lightning
                            payments. Your total lightning balance is{" "}
                            {new Intl.NumberFormat().format(
                              channels
                                ?.map((channel) =>
                                  Math.floor(channel.localBalance / 1000)
                                )
                                .reduce((a, b) => a + b, 0) || 0
                            )}{" "}
                            sats which includes{" "}
                            {new Intl.NumberFormat().format(
                              Math.floor(
                                channels
                                  ?.map((channel) =>
                                    Math.min(
                                      Math.floor(channel.localBalance / 1000),
                                      channel.unspendablePunishmentReserve
                                    )
                                  )
                                  .reduce((a, b) => a + b, 0) || 0
                              )
                            )}{" "}
                            sats reserved in your channels which cannot be
                            spent.
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
                      <>
                        <div className="text-xl font-medium balance sensitive mb-1">
                          {new Intl.NumberFormat().format(
                            Math.floor(balances.lightning.totalSpendable / 1000)
                          )}{" "}
                          sats
                        </div>
                        <FormattedFiatAmount
                          amount={balances.lightning.totalSpendable / 1000}
                        />
                      </>
                    )}
                  </CardContent>
                </div>
                <div className="flex flex-col flex-1">
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1 pr-0">
                    <CardTitle className="text-sm font-medium">
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <div className="flex flex-row gap-1 items-center justify-start text-sm font-medium">
                              Receive Limit
                              <InfoIcon className="h-3 w-3 shrink-0 text-muted-foreground" />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent className="w-[300px]">
                            Your receiving limit is the funds owned by your
                            channel partner, which will be moved to your side
                            when you receive lightning payments.
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="flex-grow pb-0">
                    {balances && (
                      <>
                        <div className="text-xl font-medium balance sensitive mb-1">
                          {new Intl.NumberFormat().format(
                            Math.floor(
                              balances.lightning.totalReceivable / 1000
                            )
                          )}{" "}
                          sats
                        </div>
                        <FormattedFiatAmount
                          amount={balances.lightning.totalReceivable / 1000}
                        />
                      </>
                    )}
                  </CardContent>
                </div>
              </CardContent>
            </Card>
            <Card className="flex flex-1 flex-col">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-2xl font-semibold">
                  On-Chain
                </CardTitle>
                <LinkIcon className="h-6 w-6 text-muted-foreground" />
              </CardHeader>
              <CardContent className="flex-grow">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1 pl-0">
                  <CardTitle className="text-sm font-medium">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger>
                          <div className="flex flex-row gap-1 items-center text-sm font-medium">
                            Balance
                            <InfoIcon className="h-3 w-3 shrink-0 text-muted-foreground" />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent className="w-[300px]">
                          Your on-chain balance can be used to open new outgoing
                          lightning channels and to ensure channels can be
                          closed when required. When channels are closed, funds
                          on your side of your channel will be returned to your
                          on-chain balance.
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
                <div>
                  {balances && (
                    <>
                      <div className="mb-1">
                        <span className="text-xl font-medium balance sensitive mb-1 mr-1">
                          {new Intl.NumberFormat().format(
                            Math.floor(balances.onchain.spendable)
                          )}{" "}
                          sats
                        </span>
                        {!!channels?.length &&
                          balances.onchain.reserved +
                            balances.onchain.spendable <
                            25_000 && (
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger>
                                  <AlertTriangleIcon className="w-3 h-3 text-muted-foreground" />
                                </TooltipTrigger>
                                <TooltipContent className="w-[300px]">
                                  You have insufficient funds in reserve to
                                  close channels or bump on-chain transactions
                                  and currently rely on the counterparty. It is
                                  recommended to deposit at least 25,000 sats to
                                  your on-chain balance.
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          )}
                      </div>
                      <FormattedFiatAmount
                        amount={balances.onchain.spendable}
                        className="mb-1"
                      />
                      {balances &&
                        balances.onchain.spendable !==
                          balances.onchain.total && (
                          <p className="text-xs text-muted-foreground animate-pulse">
                            +
                            {new Intl.NumberFormat().format(
                              balances.onchain.total -
                                balances.onchain.spendable
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

          {balances &&
            balances.onchain.pendingBalancesFromChannelClosures > 0 && (
              <Alert>
                <HourglassIcon className="h-4 w-4" />
                <AlertTitle>Pending Closed Channels</AlertTitle>
                <AlertDescription>
                  You have{" "}
                  {new Intl.NumberFormat().format(
                    balances.onchain.pendingBalancesFromChannelClosures
                  )}{" "}
                  sats pending from closed channels with
                  {[
                    ...balances.onchain.pendingBalancesDetails,
                    ...balances.onchain.pendingSweepBalancesDetails,
                  ].map((details, index) => (
                    <div key={details.channelId} className="inline">
                      &nbsp;
                      <ExternalLink
                        to={`https://amboss.space/node/${details.nodeId}`}
                        className="underline"
                      >
                        {nodes.find(
                          (node) => node.public_key === details.nodeId
                        )?.alias || "Unknown"}
                        <ExternalLinkIcon className="ml-1 w-4 h-4 inline" />
                      </ExternalLink>{" "}
                      ({new Intl.NumberFormat().format(details.amount)}{" "}
                      sats)&nbsp;
                      <ExternalLink
                        to={`${info?.mempoolUrl}/tx/${details.fundingTxId}#flow=&vout=${details.fundingTxVout}`}
                        className="underline"
                      >
                        funding tx
                        <ExternalLinkIcon className="ml-1 w-4 h-4 inline" />
                      </ExternalLink>
                      {index <
                        balances.onchain.pendingBalancesDetails.length - 1 &&
                        ","}
                    </div>
                  ))}
                  . Once spendable again these will become available in your
                  on-chain balance. Funds from channels that were force closed
                  may take up to 2 weeks to become available.{" "}
                  <ExternalLink
                    to="https://guides.getalby.com/user-guide/alby-hub/faq/why-was-my-lightning-channel-closed-and-what-to-do-next"
                    className="underline"
                  >
                    Learn more
                  </ExternalLink>
                </AlertDescription>
              </Alert>
            )}

          {channels && channels.length === 0 && (
            <EmptyState
              icon={UnplugIcon}
              title="No Channels Available"
              description="Connect to the Lightning Network by establishing your first channel and start transacting."
              buttonText="Open Channel"
              buttonLink="/channels/incoming"
            />
          )}

          <ChannelsTable
            channels={channels}
            nodes={nodes}
            longUnconfirmedZeroConfChannels={longUnconfirmedZeroConfChannels}
          />
          <ChannelsCards
            channels={channels}
            nodes={nodes}
            longUnconfirmedZeroConfChannels={longUnconfirmedZeroConfChannels}
          />
          <OnchainTransactionsTable />
        </>
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
