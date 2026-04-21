import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { ArrowDownIcon, ArrowUpIcon, LinkIcon } from "lucide-react";
import EmptyState from "src/components/EmptyState";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useInfo } from "src/hooks/useInfo";
import { useOnchainTransactions } from "src/hooks/useOnchainTransactions";
import { cn } from "src/lib/utils";
import { OnchainTransaction } from "src/types";
import { openLink } from "src/utils/openLink";

dayjs.extend(relativeTime);

type OnchainTransactionsTableProps = {
  wrapInCard?: boolean;
  title?: string;
  className?: string;
  contentClassName?: string;
  showEmptyState?: boolean;
  emptyStateTitle?: string;
  emptyStateDescription?: string;
  emptyStateButtonText?: string;
  emptyStateButtonLink?: string;
};

function typeStateLabel(tx: OnchainTransaction) {
  if (tx.type === "outgoing") {
    return tx.state === "confirmed" ? "Sent" : "Sending";
  }
  return tx.state === "confirmed" ? "Received" : "Receiving";
}

function OnchainTransactionRow({
  tx,
  mempoolUrl,
}: {
  tx: OnchainTransaction;
  mempoolUrl?: string;
}) {
  const Icon = tx.type === "outgoing" ? ArrowUpIcon : ArrowDownIcon;
  const isUnconfirmed = tx.state === "unconfirmed";
  const typeStateText = typeStateLabel(tx);

  const icon = (
    <div className="flex items-center">
      <div
        className={cn(
          "relative flex h-10 w-10 items-center justify-center rounded-full md:h-14 md:w-14",
          isUnconfirmed
            ? "bg-blue-100 dark:bg-sky-950"
            : tx.type === "outgoing"
              ? "bg-orange-100 dark:bg-amber-950"
              : "bg-green-100 dark:bg-emerald-950"
        )}
        title={`${tx.numConfirmations} confirmations`}
      >
        <Icon
          strokeWidth={3}
          className={cn(
            "size-6 md:h-8 md:w-8",
            isUnconfirmed
              ? "stroke-blue-500 dark:stroke-sky-500"
              : tx.type === "outgoing"
                ? "stroke-orange-500 dark:stroke-amber-500"
                : "stroke-green-500 dark:stroke-teal-500"
          )}
        />
      </div>
    </div>
  );

  const subtitle =
    tx.txId.length > 22
      ? `${tx.txId.slice(0, 10)}…${tx.txId.slice(-8)}`
      : tx.txId;

  return (
    <button
      type="button"
      className="transaction sensitive slashed-zero w-full cursor-pointer rounded-md p-3 text-left hover:bg-muted/50"
      onClick={() => {
        if (mempoolUrl) {
          openLink(`${mempoolUrl}/tx/${tx.txId}`);
        }
      }}
    >
      <div className={cn("flex gap-3", isUnconfirmed && "animate-pulse")}>
        {icon}
        <div className="mr-3 flex max-w-full flex-col items-start justify-center overflow-hidden text-left">
          <div className="flex items-center gap-2">
            <span className="line-clamp-1 break-all font-semibold md:text-xl">
              {typeStateText}
            </span>
            <span
              className="shrink-0 text-xs text-muted-foreground md:text-base"
              title={dayjs(tx.createdAt * 1000)
                .local()
                .format("D MMMM YYYY, HH:mm")}
            >
              {dayjs(tx.createdAt * 1000)
                .local()
                .fromNow()}
            </span>
          </div>
          <p className="line-clamp-1 break-all font-mono text-sm text-muted-foreground md:text-base">
            {subtitle}
          </p>
        </div>
        <div className="ml-auto flex shrink-0 space-x-3">
          <div className="flex flex-col items-end md:text-xl">
            <div className="flex flex-row gap-1">
              <p
                className={cn(
                  tx.type === "incoming" &&
                    "text-green-600 dark:text-emerald-500"
                )}
              >
                {tx.type === "outgoing" ? "-" : "+"}
                <FormattedBitcoinAmount
                  amount={tx.amountSat * 1000}
                  className="font-medium"
                />
              </p>
            </div>
            <FormattedFiatAmount
              className="text-xs md:text-base"
              amount={tx.amountSat}
            />
          </div>
        </div>
      </div>
    </button>
  );
}

export function OnchainTransactionsTable({
  wrapInCard = true,
  title = "On-Chain Transactions",
  className,
  contentClassName,
  showEmptyState = false,
  emptyStateTitle = "No on-chain transactions yet",
  emptyStateDescription = "Your most recent incoming and outgoing on-chain payments will show up here.",
  emptyStateButtonText = "Receive to On-chain Balance",
  emptyStateButtonLink = "/wallet/receive/onchain?type=onchain",
}: OnchainTransactionsTableProps) {
  const { data: info } = useInfo();
  const { data: transactions } = useOnchainTransactions();

  if (!transactions?.length) {
    if (!showEmptyState) {
      return null;
    }

    const emptyState = (
      <EmptyState
        icon={LinkIcon}
        title={emptyStateTitle}
        description={emptyStateDescription}
        buttonText={emptyStateButtonText}
        buttonLink={emptyStateButtonLink}
        showBorder={false}
      />
    );

    if (!wrapInCard) {
      return (
        <div className={cn("flex flex-1 flex-col", className)}>
          {emptyState}
        </div>
      );
    }

    return (
      <Card className={cn("mt-6", className)}>
        {title && (
          <CardHeader>
            <CardTitle className="text-2xl">{title}</CardTitle>
          </CardHeader>
        )}
        <CardContent className={contentClassName}>{emptyState}</CardContent>
      </Card>
    );
  }

  const rows = transactions.map((tx) => (
    <OnchainTransactionRow
      key={tx.txId}
      tx={tx}
      mempoolUrl={info?.mempoolUrl}
    />
  ));

  if (!wrapInCard) {
    return (
      <div className={cn("flex flex-1 flex-col space-y-4", className)}>
        {rows}
      </div>
    );
  }

  return (
    <Card className={cn("mt-6", className)}>
      {title && (
        <CardHeader>
          <CardTitle className="text-2xl">{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent className={cn("space-y-4", contentClassName)}>
        {rows}
      </CardContent>
    </Card>
  );
}
