import {
  AlertTriangle,
  ArrowDownIcon,
  ArrowUpIcon,
  CreditCard,
} from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import ExternalLink from "src/components/ExternalLink";
import { SelfCustodyIcon } from "src/components/icons/SelfCustodyIcon";
import Loading from "src/components/Loading";
import TransactionsList from "src/components/TransactionsList";
import { TransferFundsButton } from "src/components/TransferFundsButton";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Button } from "src/components/ui/button";
import { ALBY_HIDE_HOSTED_BALANCE_BELOW as ALBY_HIDE_HOSTED_BALANCE_LIMIT } from "src/constants.ts";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";

function Wallet() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { data: albyBalance, mutate: reloadAlbyBalance } = useAlbyBalance();

  if (!info || !balances) {
    return <Loading />;
  }

  const showMigrateCard =
    albyBalance && albyBalance.sats > ALBY_HIDE_HOSTED_BALANCE_LIMIT;

  return (
    <>
      <AppHeader title="Wallet" description="" />
      {showMigrateCard && (
        <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm p-8">
          <div className="flex flex-col items-center gap-1 text-center max-w-md">
            <SelfCustodyIcon
              width="211"
              height="64"
              viewBox="0 0 211 64"
              strokeWidth="0"
              className="text-primary-background"
            />
            <h3 className="mt-4 text-lg font-semibold">
              Your funds ({new Intl.NumberFormat().format(albyBalance.sats)}{" "}
              sats) are still hosted by Alby.
            </h3>
            <p className="text-sm text-muted-foreground mb-4">
              {channels && channels.length > 0
                ? "Transfer funds from your Alby hosted balance to your self-custodial wallet."
                : "Migrate funds from your Alby hosted balance to start using your self-custodial wallet."}
            </p>
            {channels && channels.length > 0 ? (
              <TransferFundsButton
                channels={channels}
                albyBalance={albyBalance}
                reloadAlbyBalance={reloadAlbyBalance}
              >
                Transfer Funds
              </TransferFundsButton>
            ) : (
              <Link to="/channels/first">
                <Button className="mt-4">Migrate Funds</Button>
              </Link>
            )}
          </div>
        </div>
      )}
      {hasChannelManagement && !balances.lightning.totalSpendable && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Low spending balance</AlertTitle>
          <AlertDescription>
            You won't be able to make payments until you{" "}
            <Link className="underline" to="/channels/outgoing">
              increase your spending balance.
            </Link>
          </AlertDescription>
        </Alert>
      )}
      {hasChannelManagement && !balances.lightning.totalReceivable && (
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
      <BreezRedeem />
      <div className="flex flex-col lg:flex-row justify-between lg:items-center gap-5">
        <div className="text-5xl font-semibold balance sensitive slashed-zero">
          {new Intl.NumberFormat().format(
            Math.floor(balances.lightning.totalSpendable / 1000)
          )}{" "}
          sats
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
