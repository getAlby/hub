import {
  AlertTriangleIcon,
  ArrowDownIcon,
  ArrowDownUpIcon,
  ArrowUpIcon,
  CalendarSyncIcon,
  CreditCardIcon,
} from "lucide-react";
import { Link, useSearchParams } from "react-router";
import AppHeader from "src/components/AppHeader";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import { OnchainTransactionsList } from "src/components/OnchainTransactionsList";
import TransactionsList from "src/components/TransactionsList";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LinkButton } from "src/components/ui/custom/link-button";
import { WalletActionsMenu } from "src/components/WalletActionsMenu";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useSyncWallet } from "src/hooks/useSyncWallet";

type BalanceMode = "lightning" | "onchain";

function Wallet() {
  useSyncWallet();
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances(true);
  const { data: channels } = useChannels();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeBalanceMode: BalanceMode =
    searchParams.get("mode") === "onchain" ? "onchain" : "lightning";

  const isOnchainMode = activeBalanceMode === "onchain";
  const hasChannelsOpen = !!channels?.length;

  const toggleBalanceMode = () => {
    setSearchParams(
      activeBalanceMode === "lightning" ? { mode: "onchain" } : {}
    );
  };

  if (!info || !balances) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader
        title="Wallet"
        pageTitle="Wallet"
        description=""
        contentRight={
          <div className="flex items-center gap-1 sm:gap-2">
            {hasChannelManagement && (
              <LinkButton
                to="/wallet/swap"
                variant="ghost"
                size="sm"
                className="hidden sm:inline-flex"
              >
                <ArrowDownUpIcon />
                Swap
              </LinkButton>
            )}
            <LinkButton
              to="/internal-apps/zapplanner"
              variant="ghost"
              size="sm"
              className="hidden sm:inline-flex"
            >
              <CalendarSyncIcon />
              Recurring
            </LinkButton>
            <ExternalLinkButton
              to="https://www.getalby.com/topup"
              variant="ghost"
              size="sm"
              className="hidden sm:inline-flex"
            >
              <CreditCardIcon />
              Buy
            </ExternalLinkButton>
            <WalletActionsMenu hasChannelManagement={!!hasChannelManagement} />
          </div>
        }
      />
      {!isOnchainMode &&
        hasChannelManagement &&
        hasChannelsOpen &&
        channels?.every(
          (channel) =>
            channel.localBalance < channel.unspendablePunishmentReserve * 1000
        ) && (
          <Alert>
            <AlertTriangleIcon className="h-4 w-4" />
            <AlertTitle>Channel Reserves Unmet</AlertTitle>
            <AlertDescription>
              You won't be able to make payments until you fill your channel
              reserve.{" "}
              <Link to="/channels" className="underline">
                View channel reserves
              </Link>
            </AlertDescription>
          </Alert>
        )}
      {!isOnchainMode &&
        hasChannelManagement &&
        hasChannelsOpen &&
        balances.lightning.totalReceivable <
          balances.lightning.totalSpendable * 0.1 && (
          <LowReceivingCapacityAlert />
        )}
      {!isOnchainMode &&
        hasChannelManagement &&
        channels &&
        !hasChannelsOpen && (
          <Alert>
            <AlertTriangleIcon className="h-4 w-4" />
            <AlertTitle>Open Your First Channel</AlertTitle>
            <AlertDescription className="inline">
              You won't be able to receive or send payments until you{" "}
              <Link className="underline" to="/channels/first">
                open your first channel
              </Link>
              .
            </AlertDescription>
          </Alert>
        )}
      <div className="flex w-full flex-col items-center gap-8 pt-12 pb-16 text-center">
        <div className="flex flex-col items-center gap-4">
          {hasChannelManagement ? (
            <button
              type="button"
              onClick={toggleBalanceMode}
              aria-label={`Toggle balance mode, currently ${
                isOnchainMode ? "On-chain Balance" : "Spending Balance"
              }`}
              className="inline-flex items-center justify-center gap-1 text-xs font-medium leading-none uppercase text-muted-foreground transition-colors hover:text-foreground"
            >
              {isOnchainMode ? "On-chain Balance" : "Spending Balance"}
              <ArrowDownUpIcon aria-hidden className="size-3 shrink-0" />
            </button>
          ) : (
            <span className="text-xs font-medium leading-none uppercase text-muted-foreground">
              Spending Balance
            </span>
          )}
          <div className="flex flex-col items-center gap-3">
            <div className="text-5xl md:text-6xl font-medium balance sensitive slashed-zero leading-none">
              <FormattedBitcoinAmount
                amount={
                  isOnchainMode
                    ? balances.onchain.spendable * 1000
                    : balances.lightning.totalSpendable
                }
              />
            </div>
            <FormattedFiatAmount
              className="text-3xl font-normal leading-9 text-muted-foreground"
              amount={
                isOnchainMode
                  ? balances.onchain.spendable
                  : balances.lightning.totalSpendable / 1000
              }
            />
            {isOnchainMode &&
              balances.onchain.total > balances.onchain.spendable && (
                <p className="text-sm md:text-base text-muted-foreground animate-pulse">
                  +
                  <FormattedBitcoinAmount
                    amount={
                      (balances.onchain.total - balances.onchain.spendable) *
                      1000
                    }
                  />{" "}
                  incoming
                </p>
              )}
          </div>
        </div>
        <div className="grid w-full max-w-100 grid-cols-2 items-center gap-3">
          <LinkButton
            to={
              isOnchainMode
                ? "/wallet/receive/onchain?type=onchain"
                : "/wallet/receive"
            }
            size="lg"
          >
            <ArrowDownIcon />
            Receive
          </LinkButton>
          <LinkButton to="/wallet/send" size="lg">
            <ArrowUpIcon />
            Send
          </LinkButton>
        </div>
      </div>

      {isOnchainMode ? <OnchainTransactionsList /> : <TransactionsList />}
    </>
  );
}

export default Wallet;
