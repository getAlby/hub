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
import { useOnchainTransactions } from "src/hooks/useOnchainTransactions";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import { OnchainTransaction } from "src/types";
import { openLink } from "src/utils/openLink";

dayjs.extend(relativeTime);

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
  const isPending = tx.state === "unconfirmed";
  const typeStateText = typeStateLabel(tx);
  const statusText = isPending ? "Pending" : "Confirmed";
  const createdAt = dayjs(tx.createdAt * 1000).local();

  const icon = (
    <div className="flex items-center">
      <div
        className={cn(
          "relative flex h-10 w-10 items-center justify-center rounded-full md:h-14 md:w-14",
          isPending
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
            isPending
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
          className="transaction sensitive slashed-zero w-full cursor-pointer rounded-md p-3 text-left hover:bg-muted/50"
        >
          <div className={cn("flex gap-3", isPending && "animate-pulse")}>
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
          <DialogTitle className={cn(isPending && "animate-pulse")}>
            {`${typeStateText} On-chain Transaction`}
          </DialogTitle>
          <DialogDescription>
            {isPending
              ? "This transaction is pending confirmation."
              : "This transaction has been confirmed on the blockchain."}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-6 text-sm">
          <div
            className={cn("flex items-center", isPending && "animate-pulse")}
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

          <div>
            <p>Status</p>
            <p className="text-muted-foreground">{statusText}</p>
          </div>

          <div>
            <p>Confirmations</p>
            <p className="text-muted-foreground">{tx.numConfirmations}</p>
          </div>

          <div>
            <p>Date & Time</p>
            <p className="text-muted-foreground">
              {createdAt.format("D MMMM YYYY, HH:mm")}
            </p>
          </div>

          <div>
            <p>Transaction ID</p>
            <div className="mt-1 flex items-start gap-3">
              <p className="break-all font-mono text-muted-foreground">
                {tx.txId}
              </p>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="text-muted-foreground"
                onClick={() => copyToClipboard(tx.txId)}
                aria-label="Copy transaction ID"
              >
                <CopyIcon />
              </Button>
            </div>
          </div>
        </div>
        {mempoolUrl && (
          <DialogFooter>
            <Button
              type="button"
              onClick={() => openLink(`${mempoolUrl}/tx/${tx.txId}`)}
            >
              <ExternalLinkIcon className="size-4" />
              View on Mempool
            </Button>
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  );
}

export function OnchainTransactionsTable() {
  const { data: info } = useInfo();
  const { data: transactions } = useOnchainTransactions();

  if (!transactions) {
    return null;
  }

  if (transactions.length === 0) {
    return (
      <div className="flex w-full flex-1 flex-col">
        <EmptyState
          icon={LinkIcon}
          title="No on-chain transactions yet"
          description="Your most recent incoming and outgoing on-chain transactions will show up here."
          buttonText="Receive to On-chain Balance"
          buttonLink="/wallet/receive/onchain?type=onchain"
          showBorder={false}
        />
      </div>
    );
  }

  return (
    <div className="flex w-full flex-1 flex-col space-y-4">
      {transactions.map((tx) => (
        <OnchainTransactionRow
          key={tx.txId}
          tx={tx}
          mempoolUrl={info?.mempoolUrl}
        />
      ))}
    </div>
  );
}
