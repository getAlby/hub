import {
  AlertTriangleIcon,
  ArrowDownIcon,
  ArrowDownUpIcon,
  ArrowUpIcon,
} from "lucide-react";
import { Link, useNavigate } from "react-router";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import TransactionsList from "src/components/TransactionsList";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";

export default function Lightning() {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances(true);
  const { data: channels } = useChannels();
  const navigate = useNavigate();

  if (!balances) {
    return null;
  }

  const hasChannelsOpen = !!channels?.length;
  const allChannelsBelowReserve =
    hasChannelsOpen &&
    channels?.every(
      (channel) =>
        channel.localBalanceMsat <
        channel.unspendablePunishmentReserveSat * 1000
    );
  const lowReceivingCapacity =
    hasChannelsOpen &&
    balances.lightning.totalReceivableMsat <
      balances.lightning.totalSpendableMsat * 0.1;
  const showOpenFirstChannel =
    hasChannelManagement && channels && !hasChannelsOpen;

  return (
    <>
      {hasChannelManagement && allChannelsBelowReserve && (
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
      {hasChannelManagement && lowReceivingCapacity && (
        <LowReceivingCapacityAlert />
      )}
      {showOpenFirstChannel && (
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
              onClick={() => navigate("/wallet/onchain")}
              aria-label="Toggle balance mode, currently Spending Balance"
              className="inline-flex items-center justify-center gap-1 text-xs font-medium leading-none uppercase text-muted-foreground transition-colors hover:text-foreground"
            >
              Spending Balance
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
                amountMsat={balances.lightning.totalSpendableMsat}
              />
            </div>
            <FormattedFiatAmount
              className="text-3xl font-normal leading-9 text-muted-foreground"
              amountSat={balances.lightning.totalSpendableSat}
            />
          </div>
        </div>
        <div className="grid w-full max-w-100 grid-cols-2 items-center gap-3">
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
      <TransactionsList />
    </>
  );
}
