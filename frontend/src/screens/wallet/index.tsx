import dayjs from "dayjs";
import {
  AlertTriangleIcon,
  ArrowDownIcon,
  ArrowDownUpIcon,
  ArrowUpIcon,
  CalendarSyncIcon,
  CreditCardIcon,
  ExternalLinkIcon,
  LightbulbIcon,
} from "lucide-react";
import { Link } from "react-router";
import AppHeader from "src/components/AppHeader";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import TransactionsList from "src/components/TransactionsList";
import { WalletActionsMenu } from "src/components/WalletActionsMenu";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useOnchainTransactions } from "src/hooks/useOnchainTransactions";

function Wallet() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();

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
      {hasChannelManagement &&
        !!channels?.length &&
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
      {hasChannelManagement &&
        !!channels?.length &&
        balances.lightning.totalReceivable <
          balances.lightning.totalSpendable * 0.1 && (
          <LowReceivingCapacityAlert />
        )}
      {hasChannelManagement && !channels?.length && (
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
      <div className="flex flex-col items-center gap-8 py-10 md:py-14 text-center">
        <div className="flex flex-col items-center gap-2">
          <div className="text-5xl md:text-6xl font-medium balance sensitive slashed-zero">
            <FormattedBitcoinAmount
              amount={balances.lightning.totalSpendable}
            />
          </div>
          <FormattedFiatAmount
            className="text-xl md:text-2xl"
            amount={balances.lightning.totalSpendable / 1000}
          />
        </div>
        <div className="grid w-full max-w-md grid-cols-2 items-center gap-3">
          <LinkButton to="/wallet/receive" size="lg">
            <ArrowDownIcon />
            Receive
          </LinkButton>
          <LinkButton to="/wallet/send" size="lg">
            <ArrowUpIcon />
            Send
          </LinkButton>
        </div>
      </div>

      <OnchainTransactionsAlert />
      <TransactionsList />
    </>
  );
}

export default Wallet;

function OnchainTransactionsAlert() {
  const { data: onchainTransactions } = useOnchainTransactions();
  if (
    onchainTransactions?.some(
      (tx) => dayjs().diff(tx.createdAt * 1000, "hours") < 24
    )
  ) {
    return (
      <Alert>
        <AlertTitle className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2 text-sm">
            <LightbulbIcon className="w-4 h-4" /> On-chain transactions are
            shown on the Node page
          </div>
          <LinkButton
            to="/channels"
            variant="secondary"
            size="sm"
            className="flex items-center gap-2"
          >
            <ExternalLinkIcon className="w-4 h-4" /> View On-chain transactions
          </LinkButton>
        </AlertTitle>
      </Alert>
    );
  }
  return null;
}
