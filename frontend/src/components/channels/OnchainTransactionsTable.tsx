import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  CopyIcon,
  ExternalLinkIcon,
  LinkIcon,
} from "lucide-react";
import EmptyState from "src/components/EmptyState";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { useInfo } from "src/hooks/useInfo";
import { copyToClipboard } from "src/lib/clipboard";
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

function transactionStatusLabel(tx: OnchainTransaction) {
  return tx.state === "confirmed" ? "Confirmed" : "Unconfirmed";
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
  const statusText = transactionStatusLabel(tx);
  const createdAt = dayjs(tx.createdAt * 1000).local();

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
    <Dialog>
      <DialogTrigger asChild>
        <button
          type="button"
          className="transaction sensitive slashed-zero mb-4 w-full cursor-pointer rounded-md p-3 text-left hover:bg-muted/50"
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
                  title={createdAt.format("D MMMM YYYY, HH:mm")}
                >
                  {createdAt.fromNow()}
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
      </DialogTrigger>
      <DialogContent className="slashed-zero">
        <DialogHeader>
          <DialogTitle
            className={cn(isUnconfirmed && "animate-pulse")}
          >{`${typeStateText} On-chain Transaction`}</DialogTitle>
          <DialogDescription className="max-h-[90vh] overflow-y-auto pr-2 text-start text-foreground">
            <div
              className={cn(
                "mt-6 flex items-center",
                isUnconfirmed && "animate-pulse"
              )}
            >
              {icon}
              <div className="ml-4">
                <p className="sensitive text-xl font-semibold md:text-2xl">
                  {tx.type === "outgoing" ? "-" : "+"}
                  <FormattedBitcoinAmount amount={tx.amountSat * 1000} />
                </p>
                <FormattedFiatAmount amount={tx.amountSat} />
              </div>
            </div>

            <div className="mt-8">
              <p>Status</p>
              <p className="text-muted-foreground">{statusText}</p>
            </div>

            <div className="mt-6">
              <p>Confirmations</p>
              <p className="text-muted-foreground">{tx.numConfirmations}</p>
            </div>

            <div className="mt-6">
              <p>Date & Time</p>
              <p className="text-muted-foreground">
                {createdAt.format("D MMMM YYYY, HH:mm")}
              </p>
            </div>

            <div className="mt-6">
              <p>Transaction ID</p>
              <div className="mt-1 flex items-start gap-3">
                <p className="break-all font-mono text-muted-foreground">
                  {tx.txId}
                </p>
                <button
                  type="button"
                  className="shrink-0 cursor-pointer text-muted-foreground"
                  onClick={() => copyToClipboard(tx.txId)}
                >
                  <CopyIcon className="size-4" />
                  <span className="sr-only">Copy transaction ID</span>
                </button>
              </div>
            </div>
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="sm:justify-between">
          <Button
            type="button"
            variant="outline"
            onClick={() => copyToClipboard(tx.txId)}
          >
            <CopyIcon className="size-4" />
            Copy Transaction ID
          </Button>
          <Button
            type="button"
            onClick={() => {
              if (mempoolUrl) {
                openLink(`${mempoolUrl}/tx/${tx.txId}`);
              }
            }}
          >
            <ExternalLinkIcon className="size-4" />
            View on Mempool
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
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
    return <div className={cn("flex flex-1 flex-col", className)}>{rows}</div>;
  }

  return (
    <Card className={cn("mt-6", className)}>
      {title && (
        <CardHeader>
          <CardTitle className="text-2xl">{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent className={contentClassName}>{rows}</CardContent>
    </Card>
  );
}
