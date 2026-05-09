import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc";
import {
  ArrowDownIcon,
  ArrowDownUpIcon,
  ArrowUpDownIcon,
  ArrowUpIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  TagIcon,
  XIcon,
} from "lucide-react";
import { nip19 } from "nostr-tools";
import React from "react";
import { Link } from "react-router";
import AppAvatar from "src/components/AppAvatar";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { PaymentFailedAlert } from "src/components/PaymentFailedAlert";
import PodcastingInfo from "src/components/PodcastingInfo";
import { TransactionDetailRow } from "src/components/TransactionDetailRow";
import TransactionLabels from "src/components/TransactionLabels";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { useApp } from "src/hooks/useApp";
import { useSwap } from "src/hooks/useSwaps";
import { cn, getAppDisplayName } from "src/lib/utils";
import { Transaction } from "src/types";

dayjs.extend(relativeTime);
dayjs.extend(utc);

type Props = {
  tx: Transaction;
  transactionListKey: string;
};

function safeNpubEncode(hex: string): string | undefined {
  try {
    return nip19.npubEncode(hex);
  } catch {
    return undefined;
  }
}

function safeNeventEncode(id: string): string | undefined {
  try {
    return nip19.neventEncode({
      id,
    });
  } catch {
    return undefined;
  }
}

function TransactionItem({ tx, transactionListKey }: Props) {
  const { data: app } = useApp(tx.appId);
  const swapId = tx.metadata?.swap_id;
  const { data: swap } = useSwap(swapId);
  const [showDetails, setShowDetails] = React.useState(false);
  const labels = tx.metadata?.user_labels ?? {};
  const labelEntries = Object.entries(labels);
  const type = tx.type;
  const updatedAt = dayjs(tx.updatedAt).local();

  const pubkey = tx.metadata?.nostr?.pubkey;
  const npub = pubkey ? safeNpubEncode(pubkey) : undefined;

  const payerName = tx.metadata?.payer_data?.name;
  const from =
    type === "incoming"
      ? payerName
        ? `from ${payerName}`
        : npub
          ? `zap from ${npub.substring(0, 12)}...`
          : swap
            ? `swap from ${swap.lockupAddress}`
            : undefined
      : undefined;

  const recipientIdentifier = tx.metadata?.recipient_data?.identifier;
  const to =
    type === "outgoing"
      ? npub
        ? `zap to ${npub.substring(0, 12)}...`
        : swap?.type === "out"
          ? `swap to ${swap.destinationAddress}`
          : recipientIdentifier
            ? `${tx.state === "failed" ? "payment " : ""}to ${recipientIdentifier}`
            : undefined
      : undefined;

  const eventId = tx.metadata?.nostr?.tags?.find((t) => t[0] === "e")?.[1];
  const nevent = eventId ? safeNeventEncode(eventId) : undefined;

  const bolt12Offer = tx.metadata?.offer;

  const description =
    tx.description || tx.metadata?.comment || bolt12Offer?.payer_note;

  const typeStateText =
    type == "incoming"
      ? "Received"
      : tx.state === "settled" // we only fetch settled incoming payments
        ? "Sent"
        : tx.state === "pending"
          ? "Sending"
          : "Failed";

  const Icon =
    tx.state === "failed"
      ? XIcon
      : tx.type === "outgoing"
        ? swapId
          ? ArrowUpDownIcon
          : ArrowUpIcon
        : swapId
          ? ArrowDownUpIcon
          : ArrowDownIcon;

  const typeStateIcon = (
    <div className="flex items-center">
      <div
        className={cn(
          "flex justify-center items-center rounded-full w-10 h-10 md:w-14 md:h-14 relative",
          tx.state === "failed"
            ? "bg-red-100 dark:bg-rose-950"
            : tx.state === "pending"
              ? "bg-blue-100 dark:bg-sky-950"
              : type === "outgoing"
                ? "bg-orange-100 dark:bg-amber-950"
                : "bg-green-100 dark:bg-emerald-950"
        )}
      >
        <Icon
          strokeWidth={3}
          className={cn(
            "size-6 md:w-8 md:h-8",
            tx.state === "failed"
              ? "stroke-red-500 dark:stroke-rose-500"
              : tx.state === "pending"
                ? "stroke-blue-500 dark:stroke-sky-500"
                : type === "outgoing"
                  ? "stroke-orange-500 dark:stroke-amber-500"
                  : "stroke-green-500 dark:stroke-teal-500"
          )}
        />
        {app && (
          <div
            className="absolute -bottom-1 -right-1"
            title={`${typeStateText} via ${getAppDisplayName(app.name)}`}
          >
            <AppAvatar
              app={app}
              className="border-none p-0 rounded-full w-4.5 h-4.5 md:w-6 md:h-6 shadow-xs"
            />
          </div>
        )}
      </div>
    </div>
  );

  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open) {
          setShowDetails(false);
        }
      }}
    >
      <DialogTrigger className="p-3 mb-4 hover:bg-muted/50 data-[state=open]:bg-muted cursor-pointer rounded-md slashed-zero transaction sensitive">
        <div
          className={cn(
            "flex gap-3",
            tx.state === "pending" && "animate-pulse"
          )}
        >
          {typeStateIcon}
          <div className="overflow-hidden mr-3 max-w-full text-left flex flex-col items-start justify-center">
            <div className="flex items-center gap-2">
              <span className="md:text-xl font-semibold break-all line-clamp-1">
                {typeStateText}
                {from !== undefined && <>&nbsp;{from}</>}
                {to !== undefined && <>&nbsp;{to}</>}
              </span>
              <span className="text-xs md:text-base text-muted-foreground shrink-0">
                {updatedAt.fromNow()}
              </span>
              {labelEntries.length > 0 && (
                <TagIcon
                  className="size-3 text-muted-foreground shrink-0"
                  aria-label={`${labelEntries.length} label${labelEntries.length === 1 ? "" : "s"}`}
                />
              )}
            </div>
            <p className="text-sm md:text-base text-muted-foreground break-all line-clamp-1">
              {description}
            </p>
          </div>
          <div className="flex ml-auto space-x-3 shrink-0">
            <div className="flex flex-col items-end md:text-xl">
              <div className="flex flex-row gap-1">
                <p
                  className={cn(
                    type == "incoming" && "text-green-600 dark:text-emerald-500"
                  )}
                >
                  {type == "outgoing" ? "-" : "+"}
                  <FormattedBitcoinAmount
                    amountMsat={tx.amountMsat}
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
          <DialogTitle
            className={cn(tx.state === "pending" && "animate-pulse")}
          >{`${typeStateText} Bitcoin Payment`}</DialogTitle>
        </DialogHeader>
        <div className="space-y-6 text-sm mt-2">
          <div
            className={cn(
              "flex items-center",
              tx.state === "pending" && "animate-pulse"
            )}
          >
            {typeStateIcon}
            <div className="ml-4">
              <p className="text-xl md:text-2xl font-semibold sensitive">
                {tx.type === "outgoing" ? "-" : "+"}
                <FormattedBitcoinAmount amountMsat={tx.amountMsat} />
              </p>
              <FormattedFiatAmount amountSat={tx.amountSat} />
            </div>
          </div>
          {app && (
            <TransactionDetailRow label="App">
              <Link to={`/apps/${app.id}`}>
                <p className="font-semibold text-foreground">
                  {getAppDisplayName(app.name)}
                </p>
              </Link>
            </TransactionDetailRow>
          )}
          {swapId && (
            <TransactionDetailRow label="Swap Id">
              <Link
                to={`/wallet/swap/${type === "incoming" ? "in" : "out"}/status/${swapId}`}
                className="flex items-center gap-1"
              >
                <p className="underline">{swapId}</p>
              </Link>
            </TransactionDetailRow>
          )}
          {to && <TransactionDetailRow label="To">{to}</TransactionDetailRow>}
          {payerName && (
            <TransactionDetailRow label="From">
              {payerName}
            </TransactionDetailRow>
          )}
          <TransactionDetailRow label="Date & Time">
            {updatedAt.format("D MMMM YYYY, HH:mm")}
          </TransactionDetailRow>
          {tx.state != "failed" && type == "outgoing" && (
            <TransactionDetailRow label="Fee">
              <FormattedBitcoinAmount amountMsat={tx.feesPaidMsat} />
              {tx.feesPaidMsat > 0 && (
                <>
                  &nbsp;(
                  {((tx.feesPaidMsat / tx.amountMsat) * 100).toFixed(2)}%)
                </>
              )}
            </TransactionDetailRow>
          )}
          {tx.description && (
            <TransactionDetailRow label="Description">
              {tx.description}
            </TransactionDetailRow>
          )}
          {tx.metadata?.comment && (
            <TransactionDetailRow label="Comment">
              {tx.metadata.comment}
            </TransactionDetailRow>
          )}
          {bolt12Offer?.payer_note && (
            <TransactionDetailRow label="Payer Note">
              {bolt12Offer.payer_note}
            </TransactionDetailRow>
          )}
          {/* for Alby lightning addresses the content of the zap request is
            automatically extracted and already displayed above as description */}
          {tx.metadata?.nostr && nevent && npub && (
            <TransactionDetailRow
              label={
                <ExternalLink
                  to={`https://njump.me/${nevent}`}
                  className="underline"
                >
                  Nostr Zap
                </ExternalLink>
              }
            >
              from {npub}
            </TransactionDetailRow>
          )}
          {tx.state === "failed" && (
            <div>
              <PaymentFailedAlert
                errorMessage={tx.failureReason}
                invoice={tx.invoice}
              />
            </div>
          )}
          <TransactionLabels
            id={tx.id}
            labels={labels}
            transactionListKey={transactionListKey}
          />
          <div className="w-full">
            <div
              className="flex items-center gap-2 cursor-pointer"
              onClick={() => setShowDetails(!showDetails)}
            >
              Details
              {showDetails ? (
                <ChevronUpIcon className="size-4" />
              ) : (
                <ChevronDownIcon className="size-4" />
              )}
            </div>
            {showDetails && (
              <div className="flex flex-col gap-6 mt-6">
                {tx.boostagram && <PodcastingInfo boost={tx.boostagram} />}
                {bolt12Offer && (
                  <TransactionDetailRow
                    label="BOLT-12 Offer Id"
                    copyable={bolt12Offer.id}
                  >
                    {bolt12Offer.id}
                  </TransactionDetailRow>
                )}
                {tx.preimage && (
                  <TransactionDetailRow label="Preimage" copyable={tx.preimage}>
                    {tx.preimage}
                  </TransactionDetailRow>
                )}
                <TransactionDetailRow label="Hash" copyable={tx.paymentHash}>
                  {tx.paymentHash}
                </TransactionDetailRow>
                <TransactionDetailRow label="Invoice" copyable={tx.invoice}>
                  {tx.invoice}
                </TransactionDetailRow>
                {!!tx.failureReason && (
                  <TransactionDetailRow
                    label="Failure Reason"
                    copyable={tx.failureReason}
                  >
                    {tx.failureReason}
                  </TransactionDetailRow>
                )}
                {tx.metadata && (
                  <TransactionDetailRow
                    label="Metadata"
                    copyable={JSON.stringify(tx.metadata)}
                  >
                    {JSON.stringify(tx.metadata)}
                  </TransactionDetailRow>
                )}
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export default TransactionItem;
