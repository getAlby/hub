import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc";
import { ArrowDownIcon, ArrowUpIcon, ExternalLinkIcon } from "lucide-react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { TransactionDetailRow } from "src/components/TransactionDetailRow";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { cn } from "src/lib/utils";
import { OnchainTransaction } from "src/types";

dayjs.extend(relativeTime);
dayjs.extend(utc);

type OnchainTransactionItemProps = {
  tx: OnchainTransaction;
  mempoolUrl?: string;
};

function typeStateLabel(tx: OnchainTransaction) {
  if (tx.type === "outgoing") {
    return tx.state === "confirmed" ? "Sent" : "Sending";
  }
  return tx.state === "confirmed" ? "Received" : "Receiving";
}

function OnchainTransactionItem({
  tx,
  mempoolUrl,
}: OnchainTransactionItemProps) {
  const Icon = tx.type === "outgoing" ? ArrowUpIcon : ArrowDownIcon;
  const isPending = tx.state === "unconfirmed";
  const typeStateText = typeStateLabel(tx);
  const statusText = isPending ? "Pending" : "Confirmed";
  const createdAt = dayjs(tx.createdAt * 1000).local();
  const subtitle = `${tx.txId.slice(0, 10)}…${tx.txId.slice(-8)}`;

  const icon = (
    <div className="flex items-center">
      <div
        className={cn(
          "flex items-center justify-center rounded-full h-10 w-10 md:h-14 md:w-14 relative",
          isPending
            ? "bg-blue-100 dark:bg-sky-950"
            : tx.type === "outgoing"
              ? "bg-orange-100 dark:bg-amber-950"
              : "bg-green-100 dark:bg-emerald-950"
        )}
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

  return (
    <Dialog>
      <DialogTrigger className="p-3 hover:bg-muted/50 data-[state=open]:bg-muted cursor-pointer rounded-md slashed-zero transaction sensitive">
        <div className={cn("flex gap-3", isPending && "animate-pulse")}>
          {icon}
          <div className="overflow-hidden mr-3 max-w-full text-left flex flex-col items-start justify-center">
            <div className="flex items-center gap-2">
              <span className="md:text-xl font-semibold break-all line-clamp-1">
                {typeStateText}
              </span>
              <span
                className="text-xs md:text-base text-muted-foreground shrink-0"
                title={createdAt.format("D MMMM YYYY, HH:mm")}
              >
                {createdAt.fromNow()}
              </span>
            </div>
            <p className="font-mono text-sm md:text-base text-muted-foreground break-all line-clamp-1">
              {subtitle}
            </p>
          </div>
          <div className="flex ml-auto space-x-3 shrink-0">
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
                    amountMsat={tx.amountSat * 1000}
                    className="font-medium"
                  />
                </p>
              </div>
              <FormattedFiatAmount
                className="text-xs md:text-base"
                amountSat={tx.amountSat}
              />
            </div>
          </div>
        </div>
      </DialogTrigger>
      <DialogContent className="slashed-zero max-h-[90vh]">
        <DialogHeader>
          <DialogTitle className={cn(isPending && "animate-pulse")}>
            {`${typeStateText} On-chain Transaction`}
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-6 text-sm mt-2">
          <div
            className={cn("flex items-center", isPending && "animate-pulse")}
          >
            {icon}
            <div className="ml-4">
              <p className="text-xl md:text-2xl font-semibold sensitive">
                {tx.type === "outgoing" ? "-" : "+"}
                <FormattedBitcoinAmount amountMsat={tx.amountSat * 1000} />
              </p>
              <FormattedFiatAmount amountSat={tx.amountSat} />
            </div>
          </div>

          <TransactionDetailRow label="Status">
            {statusText}
          </TransactionDetailRow>
          <TransactionDetailRow label="Confirmations">
            {tx.numConfirmations}
          </TransactionDetailRow>
          <TransactionDetailRow label="Date & Time">
            {createdAt.format("D MMMM YYYY, HH:mm")}
          </TransactionDetailRow>
          <TransactionDetailRow label="Transaction ID" copyable={tx.txId}>
            {tx.txId}
          </TransactionDetailRow>
        </div>
        <DialogFooter>
          <ExternalLinkButton to={`${mempoolUrl}/tx/${tx.txId}`}>
            <ExternalLinkIcon className="size-4" />
            View on Mempool
          </ExternalLinkButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default OnchainTransactionItem;
