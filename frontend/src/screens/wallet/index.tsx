import {
  AlertTriangleIcon,
  ArrowDownIcon,
  ArrowDownUpIcon,
  ArrowUpIcon,
  CreditCardIcon,
} from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import ResponsiveButton from "src/components/ResponsiveButton";
import TransactionsList from "src/components/TransactionsList";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Button } from "src/components/ui/button";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";

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
        description=""
        contentRight={
          <>
            <Link to="/wallet/receive">
              <ResponsiveButton icon={ArrowDownIcon} text="Receive" size="lg" />
            </Link>
            <Link to="/wallet/send">
              <ResponsiveButton icon={ArrowUpIcon} text="Send" size="lg" />
            </Link>
          </>
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
          <AlertDescription>
            You won't be able to receive or send payments until you{" "}
            <Link className="underline" to="/channels/first">
              open your first channel
            </Link>
            .
          </AlertDescription>
        </Alert>
      )}
      <div className="flex flex-col xl:flex-row justify-between xl:items-start gap-5">
        <div className="flex flex-col gap-1 text-center xl:text-left">
          <div className="text-5xl font-medium balance sensitive slashed-zero">
            {new Intl.NumberFormat().format(
              Math.floor(balances.lightning.totalSpendable / 1000)
            )}{" "}
            sats
          </div>
          <FormattedFiatAmount
            className="text-xl"
            amount={balances.lightning.totalSpendable / 1000}
          />
        </div>
        <div className="grid grid-cols-2 items-center gap-3">
          <ExternalLink to="https://www.getalby.com/topup">
            <Button className="w-full" variant="secondary">
              <CreditCardIcon className="h-4 w-4 shrink-0 mr-2" />
              Buy Bitcoin
            </Button>
          </ExternalLink>
          {hasChannelManagement && (
            <Link to="/wallet/swap">
              <Button className="w-full" variant="secondary">
                <ArrowDownUpIcon className="h-4 w-4 shrink-0 mr-2" />
                Swap
              </Button>
            </Link>
          )}
        </div>
      </div>

      <TransactionsList />
    </>
  );
}

export default Wallet;
