import dayjs from "dayjs";
import {
  AlertTriangleIcon,
  ArrowDownUpIcon,
  ArrowRightIcon,
  CopyIcon,
  HeartIcon,
  InfoIcon,
  LinkIcon,
  Settings2Icon,
  UnplugIcon,
  ZapIcon,
} from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router";
import AppHeader from "src/components/AppHeader.tsx";
import { ChannelsCards } from "src/components/channels/ChannelsCards.tsx";
import { ChannelsTable } from "src/components/channels/ChannelsTable.tsx";
import { HealthCheckAlert } from "src/components/channels/HealthcheckAlert";
import { LDKChannelMonitorSizeAlert } from "src/components/channels/LDKChannelMonitorSizeAlert";
import { LDKChannelWithoutPeerAlert } from "src/components/channels/LDKChannelWithoutPeerAlert";
import EmptyState from "src/components/EmptyState.tsx";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import { PendingClosedChannelsAlert } from "src/components/PendingClosedChannelsAlert";
import ResponsiveButton from "src/components/ResponsiveButton";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card.tsx";
import CircleProgress from "src/components/ui/custom/circle-progress";
import { LinkButton } from "src/components/ui/custom/link-button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu.tsx";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip.tsx";
import { ProDropdownMenuItem } from "src/components/UpgradeDialog";
import { ONCHAIN_DUST_SATS } from "src/constants.ts";
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
  MempoolTransaction,
} from "src/types";
import { request } from "src/utils/request";

export default function Channels() {
  useSyncWallet();
  const { data: channels } = useChannels();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances(true);
  const navigate = useNavigate();
  const [longUnconfirmedZeroConfChannels, setLongUnconfirmedZeroConfChannels] =
    React.useState<LongUnconfirmedZeroConfChannel[]>([]);

  const nodeHealth = channels ? getNodeHealth(channels) : 0;

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
      } catch {
        _longUnconfirmedZeroConfChannels.push({
          id: channel.id,
          message: "Channel transaction not in the mempool yet",
        });
      }
    }
    setLongUnconfirmedZeroConfChannels(_longUnconfirmedZeroConfChannels);
  }, [channels]);

  React.useEffect(() => {
    findUnconfirmedChannels();
  }, [findUnconfirmedChannels]);

  return (
    <>
      <AppHeader
        title="Node"
        pageTitle="Node"
        contentRight={
          hasChannelManagement && (
            <div className="flex gap-3 items-center justify-center">
              <DropdownMenu modal={false}>
                <ResponsiveButton
                  asChild
                  icon={Settings2Icon}
                  text="Advanced"
                  variant="secondary"
                >
                  <DropdownMenuTrigger />
                </ResponsiveButton>
                <DropdownMenuContent className="w-56">
                  <DropdownMenuGroup>
                    <DropdownMenuLabel>Node</DropdownMenuLabel>
                    <DropdownMenuItem>
                      <div
                        className="flex flex-row gap-2 items-center w-full cursor-pointer"
                        onClick={() => {
                          if (!nodeConnectionInfo) {
                            return;
                          }
                          copyToClipboard(nodeConnectionInfo.pubkey);
                        }}
                      >
                        <div>Public key</div>
                        <div className="overflow-hidden text-ellipsis flex-1 text-muted-foreground text-xs">
                          {nodeConnectionInfo?.pubkey || "Loading..."}
                        </div>
                        {nodeConnectionInfo && (
                          <CopyIcon className="shrink-0 size-4" />
                        )}
                      </div>
                    </DropdownMenuItem>
                    {nodeConnectionInfo?.address &&
                      nodeConnectionInfo?.port && (
                        <DropdownMenuItem>
                          <div
                            className="flex flex-row gap-2 items-center w-full cursor-pointer"
                            onClick={() => {
                              const connectionAddress = `${nodeConnectionInfo.pubkey}@${nodeConnectionInfo.address}:${nodeConnectionInfo.port}`;
                              copyToClipboard(connectionAddress);
                            }}
                          >
                            <div>URI</div>
                            <div className="overflow-hidden text-ellipsis flex-1 text-muted-foreground text-xs">
                              {nodeConnectionInfo.pubkey.substring(0, 6)}...@
                              {nodeConnectionInfo.address}:
                              {nodeConnectionInfo.port}
                            </div>
                            <CopyIcon className="shrink-0 size-4" />
                          </div>
                        </DropdownMenuItem>
                      )}
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuLabel>On-Chain</DropdownMenuLabel>
                    <DropdownMenuItem asChild>
                      <Link to="onchain/buy-bitcoin" className="w-full">
                        Buy
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link to="/channels/onchain/deposit-bitcoin">
                        Deposit
                      </Link>
                    </DropdownMenuItem>
                    {(balances?.onchain.spendableSat || 0) >
                      ONCHAIN_DUST_SATS && (
                      <DropdownMenuItem
                        onClick={() => navigate("/wallet/withdraw")}
                      >
                        Withdraw
                      </DropdownMenuItem>
                    )}
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  {hasChannelManagement && (
                    <DropdownMenuGroup>
                      <DropdownMenuLabel>Swaps</DropdownMenuLabel>
                      <DropdownMenuItem
                        onClick={() => navigate("/wallet/swap?type=in")}
                        className="cursor-pointer"
                      >
                        <div className="mr-2 text-muted-foreground flex flex-row items-center">
                          <LinkIcon className="size-4" />
                          <ArrowRightIcon className="size-4" />
                          <ZapIcon className="size-4" />
                        </div>
                        Swap in
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => navigate("/wallet/swap?type=out")}
                        className="cursor-pointer"
                      >
                        <div className="mr-2 text-muted-foreground flex flex-row items-center">
                          <ZapIcon className="size-4" />
                          <ArrowRightIcon className="size-4" />
                          <LinkIcon className="size-4" />
                        </div>
                        Swap out
                      </DropdownMenuItem>
                    </DropdownMenuGroup>
                  )}
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
                    {info?.backendType === "LDK" && (
                      <ProDropdownMenuItem
                        onClick={() => navigate("/wallet/node-alias")}
                      >
                        Set Node Alias
                      </ProDropdownMenuItem>
                    )}
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>

              <LinkButton
                to="/wallet/swap"
                variant="secondary"
                className="hidden sm:flex"
              >
                <ArrowDownUpIcon />
                Swap
              </LinkButton>
              <LinkButton to="/channels/incoming">Open Channel</LinkButton>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/node/node-health">
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
                          className="size-4"
                          stroke={"var(--color-primary)"}
                          strokeWidth={3}
                          fill={
                            nodeHealth === 100
                              ? "var(--color-primary)"
                              : "transparent"
                          }
                        />
                      </CircleProgress>
                    </ExternalLink>
                  </TooltipTrigger>
                  <TooltipContent>Node health: {nodeHealth}%</TooltipContent>
                </Tooltip>
              </TooltipProvider>
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
                  channel.remoteBalanceMsat <
                  (channel.localBalanceMsat + channel.remoteBalanceMsat) * 0.2
              ) && <LowReceivingCapacityAlert />}
            </>
          )}

          <div
            className={cn(
              "flex flex-col sm:flex-row flex-wrap gap-3 slashed-zero"
            )}
          >
            <Card className="flex flex-1 sm:flex-2 flex-col">
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
                          <TooltipContent>
                            Your spending balance is the funds on your side of
                            your channels, which you can use to make lightning
                            payments. Your total lightning balance is{" "}
                            <FormattedBitcoinAmount
                              amountMsat={
                                channels
                                  ?.map((channel) => channel.localBalanceMsat)
                                  .reduce((a, b) => a + b, 0) || 0
                              }
                            />{" "}
                            which includes{" "}
                            <FormattedBitcoinAmount
                              amountMsat={
                                channels
                                  ?.map((channel) =>
                                    Math.min(
                                      channel.localBalanceMsat,
                                      channel.unspendablePunishmentReserveSat *
                                        1000
                                    )
                                  )
                                  .reduce((a, b) => a + b, 0) || 0
                              }
                            />{" "}
                            reserved in your channels which cannot be spent.
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="grow pb-0">
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
                          <FormattedBitcoinAmount
                            amountMsat={balances.lightning.totalSpendableMsat}
                          />
                        </div>
                        <FormattedFiatAmount
                          amountSat={balances.lightning.totalSpendableSat}
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
                          <TooltipContent>
                            Your receiving limit is the funds owned by your
                            channel partner, which will be moved to your side
                            when you receive lightning payments.
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="grow pb-0">
                    {balances && (
                      <>
                        <div className="text-xl font-medium balance sensitive mb-1">
                          <FormattedBitcoinAmount
                            amountMsat={balances.lightning.totalReceivableMsat}
                          />
                        </div>
                        <FormattedFiatAmount
                          amountSat={balances.lightning.totalReceivableSat}
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
              <CardContent className="grow">
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
                        <TooltipContent>
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
                        <span className="mr-1 text-xl font-medium balance sensitive">
                          <FormattedBitcoinAmount
                            amountMsat={balances.onchain.spendableSat * 1000}
                          />
                        </span>
                        {!!channels?.length &&
                          balances.onchain.reservedSat +
                            balances.onchain.spendableSat <
                            25_000 && (
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger>
                                  <AlertTriangleIcon className="size-3 text-muted-foreground" />
                                </TooltipTrigger>
                                <TooltipContent>
                                  You have insufficient funds in reserve to
                                  close channels or bump on-chain transactions
                                  and currently rely on the counterparty. It is
                                  recommended to deposit at least{" "}
                                  <FormattedBitcoinAmount
                                    amountMsat={25_000 * 1000}
                                  />{" "}
                                  to your on-chain balance.
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          )}
                      </div>
                      <FormattedFiatAmount
                        amountSat={balances.onchain.spendableSat}
                        className="mb-1"
                      />
                      {balances.onchain.totalSat >
                        balances.onchain.spendableSat && (
                        <p className="text-xs text-muted-foreground animate-pulse">
                          +
                          <FormattedBitcoinAmount
                            amountMsat={
                              (balances.onchain.totalSat -
                                balances.onchain.spendableSat) *
                              1000
                            }
                          />{" "}
                          incoming
                        </p>
                      )}
                    </>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>

          {balances && (
            <PendingClosedChannelsAlert balance={balances.onchain} />
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

          <LDKChannelMonitorSizeAlert />
          <LDKChannelWithoutPeerAlert />

          <ChannelsTable
            channels={channels}
            longUnconfirmedZeroConfChannels={longUnconfirmedZeroConfChannels}
          />
          <ChannelsCards
            channels={channels}
            longUnconfirmedZeroConfChannels={longUnconfirmedZeroConfChannels}
          />
        </>
      )}
    </>
  );
}

function getNodeHealth(channels: Channel[]) {
  const totalChannelCapacitySat = channels
    .map(
      (channel) => (channel.localBalanceMsat + channel.remoteBalanceMsat) / 1000
    )
    .reduce((a, b) => a + b, 0);
  const averageChannelBalance =
    channels
      .map((channel) => {
        const totalBalanceMsat =
          channel.localBalanceMsat + channel.remoteBalanceMsat;
        const expectedBalanceMsat = totalBalanceMsat / 2;
        const actualBalance =
          Math.min(channel.localBalanceMsat, channel.remoteBalanceMsat) /
          expectedBalanceMsat;
        return actualBalance;
      })
      .reduce((a, b) => a + b, 0) / (channels.length || 1);

  const numUniqueChannelPartners = new Set(
    channels.map((channel) => channel.remotePubkey)
  ).size;

  const nodeHealth = Math.ceil(
    numUniqueChannelPartners *
      (100 / 2) * // 2 or more channels is great
      (Math.min(totalChannelCapacitySat, 1_000_000) / 1_000_000) * // 1 million sats or more is great
      (0.9 + averageChannelBalance * 0.1) // +10% for perfectly balanced channels
  );

  if (nodeHealth > 95) {
    // prevent OCD
    return 100;
  }

  return nodeHealth;
}
