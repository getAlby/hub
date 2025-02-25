import {
  AlertTriangle,
  ArrowDownIcon,
  ArrowLeftRight,
  ArrowUpIcon,
  CreditCard,
} from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import TransactionsList from "src/components/TransactionsList";
import { TransferFundsButton } from "src/components/TransferFundsButton";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Button, LinkButton } from "src/components/ui/button";
import { ALBY_HIDE_HOSTED_BALANCE_BELOW as ALBY_HIDE_HOSTED_BALANCE_LIMIT } from "src/constants.ts";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useTransactions } from "src/hooks/useTransactions";

function Wallet() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances, mutate: reloadBalances } = useBalances();
  const { data: channels } = useChannels();
  const { data: albyBalance, mutate: reloadAlbyBalance } = useAlbyBalance();
  const { mutate: reloadTransactions } = useTransactions();

  if (!info || !balances) {
    return <Loading />;
  }

  const showMigrateCard =
    albyBalance && albyBalance.sats > ALBY_HIDE_HOSTED_BALANCE_LIMIT;
  const needsChannels = hasChannelManagement && channels && channels.length < 1;

  return (
    <>
      <AppHeader title="Wallet" description="" />
      {showMigrateCard && (
        <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm p-8">
          <div className="flex flex-col items-center gap-1 text-center max-w-md">
            <ArrowLeftRight className="w-10 h-10 text-primary-background" />
            <h3 className="mt-4 text-lg font-semibold">
              You still have{" "}
              <span className="font-bold slashed-zero sensitive">
                {new Intl.NumberFormat().format(albyBalance.sats)}
              </span>{" "}
              sats in your Alby shared wallet
            </h3>
            <p className="text-sm text-muted-foreground mb-4">
              Transfer funds from your Alby hosted balance.
            </p>
            {needsChannels ? (
              <LinkButton to="/channels/first">Transfer Funds</LinkButton>
            ) : (
              <TransferFundsButton
                channels={channels}
                albyBalance={albyBalance}
                onTransferComplete={() =>
                  Promise.all([
                    reloadAlbyBalance(),
                    reloadBalances(),
                    reloadTransactions(),
                  ])
                }
              >
                Transfer Funds
              </TransferFundsButton>
            )}
          </div>
        </div>
      )}
      {hasChannelManagement &&
        !!channels?.length &&
        channels?.every(
          (channel) =>
            channel.localBalance < channel.unspendablePunishmentReserve * 1000
        ) &&
        !showMigrateCard && (
          <Alert>
            <AlertTriangle className="h-4 w-4" />
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
        !balances.lightning.totalReceivable &&
        !showMigrateCard && (
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Low receiving capacity</AlertTitle>
            <AlertDescription>
              You won't be able to receive payments until you{" "}
              <Link className="underline" to="/channels/incoming">
                increase your receiving capacity.
              </Link>
            </AlertDescription>
          </Alert>
        )}
      {hasChannelManagement && !channels?.length && !showMigrateCard && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Open Your First Channel</AlertTitle>
          <AlertDescription>
            You won't be able to receive or send payments until you{" "}
            <Link className="underline" to="/channels/first">
              open your first channel
            </Link>
            .
          </AlertDescription>
        </Alert>
      )}
      <BreezRedeem />
      <div className="flex flex-col xl:flex-row justify-between xl:items-center gap-5">
        <div className="flex flex-col gap-1 text-center xl:text-left">
          <div className="text-5xl font-semibold balance sensitive slashed-zero">
            {new Intl.NumberFormat().format(
              Math.floor(balances.lightning.totalSpendable / 1000)
            )}{" "}
            sats
          </div>
          <FormattedFiatAmount
            className="text-xl font-semibold"
            amount={balances.lightning.totalSpendable / 1000}
          />
        </div>
        <div className="grid grid-cols-2 items-center gap-4 sm:grid-cols-3">
          <ExternalLink
            to="https://www.getalby.com/topup"
            className="col-span-2 sm:col-span-1"
          >
            <Button size="lg" className="w-full" variant="secondary">
              <CreditCard className="h-4 w-4 shrink-0 mr-2" />
              Buy Bitcoin
            </Button>
          </ExternalLink>
          <Link to="/wallet/receive">
            <Button size="lg" className="w-full">
              <ArrowDownIcon className="h-4 w-4 shrink-0 mr-2" />
              Receive
            </Button>
          </Link>
          <Link to="/wallet/send">
            <Button size="lg" className="w-full">
              <ArrowUpIcon className="h-4 w-4 shrink-0 mr-2" />
              Send
            </Button>
          </Link>
        </div>
      </div>

      <TransactionsList />
    </>
  );
}

export default Wallet;
