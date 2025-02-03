import {
  AlertTriangle,
  ArrowBigRightDash,
  ArrowRight,
  ChevronDown,
  CopyIcon,
  ExternalLinkIcon,
  Heart,
  HourglassIcon,
  InfoIcon,
  LinkIcon,
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
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
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
  CardHeader,
  CardTitle,
} from "src/components/ui/card.tsx";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu.tsx";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
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
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { copyToClipboard } from "src/lib/clipboard.ts";
import { cn } from "src/lib/utils.ts";
import { Channel, CreateInvoiceRequest, Node, Transaction } from "src/types";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request";

export default function Channels() {
  useSyncWallet();
  const { data: channels, mutate: reloadChannels } = useChannels();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { data: balances, mutate: reloadBalances } = useBalances();
  const { data: albyBalance, mutate: reloadAlbyBalance } = useAlbyBalance();
  const navigate = useNavigate();
  const [nodes, setNodes] = React.useState<Node[]>([]);
  const [swapInAmount, setSwapInAmount] = React.useState("");
  const [swapOutAmount, setSwapOutAmount] = React.useState("");
  const [swapOutDialogOpen, setSwapOutDialogOpen] = React.useState(false);
  const [swapInDialogOpen, setSwapInDialogOpen] = React.useState(false);
  const [loadingSwap, setLoadingSwap] = React.useState(false);
  const { getNewAddress } = useOnchainAddress();

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

  function openSwapOutDialog() {
    setSwapOutAmount(
      Math.floor(
        ((findChannelWithLargestBalance("localSpendableBalance")
          ?.localSpendableBalance || 0) *
          0.9) /
          1000
      ).toString()
    );
    setSwapOutDialogOpen(true);
  }
  function openSwapInDialog() {
    setSwapInAmount(
      Math.floor(
        ((findChannelWithLargestBalance("remoteBalance")?.remoteBalance || 0) *
          0.9) /
          1000
      ).toString()
    );
    setSwapInDialogOpen(true);
  }

  function findChannelWithLargestBalance(
    balanceType: "remoteBalance" | "localSpendableBalance"
  ): Channel | undefined {
    if (!channels || channels.length === 0) {
      return undefined;
    }

    return channels.reduce((prevLargest, current) => {
      return current[balanceType] > prevLargest[balanceType]
        ? current
        : prevLargest;
    }, channels[0]);
  }

  const showHostedBalance =
    albyBalance && albyBalance.sats > ALBY_HIDE_HOSTED_BALANCE_LIMIT;

  return (
    <>
      <AppHeader
        title="Node"
        description="Manage your lightning node liquidity."
        contentRight={
          <div className="flex gap-3 items-center justify-center">
            {/* TODO: move these dialogs to a new file */}
            <Dialog
              open={swapOutDialogOpen}
              onOpenChange={setSwapOutDialogOpen}
            >
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Swap out funds</DialogTitle>
                  <DialogDescription>
                    Funds from one of your channels will be sent to your
                    on-chain balance via a swap service. This helps restore your
                    inbound liquidity.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid grid-cols-4">
                  <Label className="pt-3">Amount (sats)</Label>
                  <div className="col-span-3">
                    <Input
                      value={swapOutAmount}
                      onChange={(e) => setSwapOutAmount(e.target.value)}
                    />
                    <p className="text-muted-foreground text-xs p-2">
                      The amount is set to 90% of the maximum spending capacity
                      available in one of your lightning channels.
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <LoadingButton
                    loading={loadingSwap}
                    type="submit"
                    onClick={async () => {
                      setLoadingSwap(true);
                      const onchainAddress = await getNewAddress();
                      if (onchainAddress) {
                        openLink(
                          `https://boltz.exchange/?sendAsset=LN&receiveAsset=BTC&sendAmount=${swapOutAmount}&destination=${onchainAddress}&ref=alby`
                        );
                      }
                      setLoadingSwap(false);
                    }}
                  >
                    Swap out
                  </LoadingButton>
                </DialogFooter>
              </DialogContent>
            </Dialog>
            <Dialog open={swapInDialogOpen} onOpenChange={setSwapInDialogOpen}>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Swap in funds</DialogTitle>
                  <DialogDescription>
                    Swap on-chain funds into your lightning channels via a swap
                    service, increasing your spending balance using on-chain
                    funds from{" "}
                    <Link to="/wallet/withdraw" className="underline">
                      your hub
                    </Link>{" "}
                    or an external wallet.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid grid-cols-4">
                  <Label className="pt-3">Amount (sats)</Label>
                  <div className="col-span-3">
                    <Input
                      value={swapInAmount}
                      onChange={(e) => setSwapInAmount(e.target.value)}
                    />
                    <p className="text-muted-foreground text-xs p-2">
                      The amount is set to 90% of the maximum receiving capacity
                      available in one of your lightning channels.
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <LoadingButton
                    loading={loadingSwap}
                    type="submit"
                    onClick={async () => {
                      setLoadingSwap(true);
                      try {
                        const transaction = await request<Transaction>(
                          "/api/invoices",
                          {
                            method: "POST",
                            headers: {
                              "Content-Type": "application/json",
                            },
                            body: JSON.stringify({
                              amount: parseInt(swapInAmount) * 1000,
                              description: "Boltz Swap In",
                            } as CreateInvoiceRequest),
                          }
                        );
                        if (!transaction) {
                          throw new Error("no transaction in response");
                        }
                        openLink(
                          `https://boltz.exchange/?sendAsset=BTC&receiveAsset=LN&sendAmount=${swapInAmount}&destination=${transaction.invoice}&ref=alby`
                        );
                      } catch (error) {
                        toast({
                          variant: "destructive",
                          title: "Failed to generate swap invoice",
                          description: "" + error,
                        });
                      }
                      setLoadingSwap(false);
                    }}
                  >
                    Swap in
                  </LoadingButton>
                </DialogFooter>
              </DialogContent>
            </Dialog>
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
                  <DropdownMenuLabel>Swaps</DropdownMenuLabel>
                  <DropdownMenuItem
                    onClick={openSwapInDialog}
                    className="cursor-pointer"
                  >
                    <div className="mr-2 text-muted-foreground flex flex-row items-center">
                      <LinkIcon className="w-4 h-4" />
                      <ArrowRight className="w-4 h-4" />
                      <ZapIcon className="w-4 h-4" />
                    </div>
                    Swap in
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={openSwapOutDialog}
                    className="cursor-pointer"
                  >
                    <div className="mr-2 text-muted-foreground flex flex-row items-center">
                      <ZapIcon className="w-4 h-4" />
                      <ArrowRight className="w-4 h-4" />
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
              <AlertTitle>Low receiving limit</AlertTitle>
              <AlertDescription>
                You likely won't be able to receive payments until you{" "}
                <Link className="underline" to="/channels/incoming">
                  increase your receiving limits.
                </Link>
              </AlertDescription>
            </Alert>
          )}
        </>
      )}

      <div
        className={cn("flex flex-col sm:flex-row flex-wrap gap-3 slashed-zero")}
      >
        <Card className="flex flex-1 sm:flex-[2] flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="font-semibold text-2xl">Lightning</CardTitle>
            <ZapIcon className="h-6 w-6 text-muted-foreground" />
          </CardHeader>

          <CardContent className="flex flex-col sm:flex-row pl-0 flex-wrap">
            {showHostedBalance && (
              <div className="flex flex-col flex-1">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1 pr-0">
                  <CardTitle className="text-sm font-medium">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger>
                          <div className="flex flex-row gap-1 items-center justify-start text-sm font-medium">
                            Alby Hosted Balance
                            <InfoIcon className="h-3 w-3 shrink-0 text-muted-foreground" />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent className="w-[300px]">
                          These are the funds from your shared Alby wallet,
                          which will be migrated to your Hub spending balance
                          upon transfer.
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex-grow pb-0">
                  <div className="flex flex-row gap-6 items-center justify-between md:justify-start">
                    <div className="flex-flex-col gap-1">
                      <div className="text-xl font-medium">
                        {new Intl.NumberFormat().format(albyBalance.sats)} sats
                      </div>
                      <FormattedFiatAmount amount={albyBalance.sats} />
                    </div>
                    <TransferFundsButton
                      variant="secondary"
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
                      Transfer <ArrowBigRightDash />
                    </TransferFundsButton>
                  </div>
                </CardContent>
              </div>
            )}
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
                        Your receiving limit is the funds owned by your channel
                        partner, which will be moved to your side when you
                        receive lightning payments.
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
                        Math.floor(balances.lightning.totalReceivable / 1000)
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
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1">
            <CardTitle className="text-2xl font-semibold">On-Chain</CardTitle>
            <LinkIcon className="h-6 w-6 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex-grow">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 pl-0">
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
            <div className="text-2xl balance sensitive">
              {balances && (
                <>
                  <div className="text-xl font-medium balance sensitive mb-1">
                    {new Intl.NumberFormat().format(
                      Math.floor(balances.onchain.spendable)
                    )}{" "}
                    sats
                  </div>
                  <FormattedFiatAmount
                    amount={balances.onchain.spendable}
                    className="mb-1"
                  />
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
            sats pending from closed channels with
            {balances.onchain.pendingBalancesDetails.map((details, index) => (
              <div key={details.channelId} className="inline">
                &nbsp;
                <ExternalLink
                  to={`https://amboss.space/node/${details.nodeId}`}
                  className="underline"
                >
                  {nodes.find((node) => node.public_key === details.nodeId)
                    ?.alias || "Unknown"}
                  <ExternalLinkIcon className="ml-1 w-4 h-4 inline" />
                </ExternalLink>{" "}
                ({new Intl.NumberFormat().format(details.amount)} sats)&nbsp;
                <ExternalLink
                  to={`https://mempool.space/tx/${details.fundingTxId}#flow=&vout=${details.fundingTxVout}`}
                  className="underline"
                >
                  funding tx
                  <ExternalLinkIcon className="ml-1 w-4 h-4 inline" />
                </ExternalLink>
                {index < balances.onchain.pendingBalancesDetails.length - 1 &&
                  ","}
              </div>
            ))}
            . Once spendable again these will become available in your on-chain
            balance. Funds from channels that were force closed may take up to 2
            weeks to become available.{" "}
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
