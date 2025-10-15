import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import {
  ArrowDownIcon,
  ArrowDownUpIcon,
  ArrowUpDownIcon,
  ArrowUpIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  CopyIcon,
  XIcon,
} from "lucide-react";
import { nip19 } from "nostr-tools";
import React from "react";
import { Link } from "react-router-dom";
import AppAvatar from "src/components/AppAvatar";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { PaymentFailedAlert } from "src/components/PaymentFailedAlert";
import PodcastingInfo from "src/components/PodcastingInfo";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { ALBY_ACCOUNT_APP_NAME } from "src/constants";
import { useApp } from "src/hooks/useApp";
import { useSwap } from "src/hooks/useSwaps";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import { Transaction } from "src/types";

dayjs.extend(utc);

type Props = {
  tx: Transaction;
};

function safeNpubEncode(hex: string): string | undefined {
  try {
    return nip19.npubEncode(hex);
  } catch {
    return undefined;
  }
}

function TransactionItem({ tx }: Props) {
  const { data: app } = useApp(tx.appId);
  const swapId = tx.metadata?.swap_id;
  const { data: swap } = useSwap(swapId);
  const [showDetails, setShowDetails] = React.useState(false);
  const type = tx.type;

  const amountSatsWithFees = Math.floor(
    (tx.amount + (tx.feesPaid || 0)) / 1000
  );

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

  const copy = (text: string) => {
    copyToClipboard(text);
  };

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
            title={`${typeStateText} via ${app.name === ALBY_ACCOUNT_APP_NAME ? "Alby Account" : app.name}`}
          >
            <AppAvatar
              app={app}
              className="border-none p-0 rounded-full w-[18px] h-[18px] md:w-6 md:h-6 shadow-xs"
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
      <DialogTrigger className="p-3 mb-4 hover:bg-muted/50 data-[state=selected]:bg-muted cursor-pointer rounded-md slashed-zero transaction sensitive">
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
                {dayjs(tx.updatedAt).fromNow()}
              </span>
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
                  <span className="font-medium">
                    {new Intl.NumberFormat().format(amountSatsWithFees)}
                  </span>
                </p>
                <p className="text-muted-foreground">
                  {amountSatsWithFees == 1 ? "sat" : "sats"}
                </p>
              </div>
              <FormattedFiatAmount
                className="text-xs md:text-base"
                amount={amountSatsWithFees}
              />
            </div>
          </div>
        </div>
      </DialogTrigger>
      <DialogContent className="slashed-zero">
        <DialogHeader>
          <DialogTitle
            className={cn(tx.state === "pending" && "animate-pulse")}
          >{`${typeStateText} Bitcoin Payment`}</DialogTitle>
          <DialogDescription className="text-start text-foreground max-h-[90vh] overflow-y-auto pr-2">
            <div
              className={cn(
                "flex items-center mt-6",
                tx.state === "pending" && "animate-pulse"
              )}
            >
              {typeStateIcon}
              <div className="ml-4">
                <p className="text-xl md:text-2xl font-semibold sensitive">
                  {new Intl.NumberFormat().format(amountSatsWithFees)}{" "}
                  {amountSatsWithFees == 1 ? "sat" : "sats"}
                </p>
                <FormattedFiatAmount amount={amountSatsWithFees} />
              </div>
            </div>
            {app && (
              <div className="mt-8">
                <p>App</p>
                <Link to={`/apps/${app.id}`}>
                  <p className="font-semibold">
                    {app.name === ALBY_ACCOUNT_APP_NAME
                      ? "Alby Account"
                      : app.name}
                  </p>
                </Link>
              </div>
            )}
            {swapId && (
              <div className="mt-8">
                <p>Swap Id</p>
                <Link
                  to={`/wallet/swap/${type === "incoming" ? "in" : "out"}/status/${swapId}`}
                  className="flex items-center gap-1"
                >
                  <p className="underline">{swapId}</p>
                </Link>
              </div>
            )}
            {to && (
              <div className="mt-6">
                <p>To</p>
                <p className="text-muted-foreground">{to}</p>
              </div>
            )}
            {payerName && (
              <div className="mt-6">
                <p>From</p>
                <p className="text-muted-foreground">{payerName}</p>
              </div>
            )}
            <div className="mt-6">
              <p>Date & Time</p>
              <p className="text-muted-foreground">
                {dayjs(tx.updatedAt).local().format("D MMMM YYYY, HH:mm")}
              </p>
            </div>
            {tx.state != "failed" && type == "outgoing" && (
              <div className="mt-6">
                <p>Amount</p>
                <p className="text-muted-foreground">
                  {new Intl.NumberFormat().format(Math.floor(tx.amount / 1000))}{" "}
                  {Math.floor(tx.amount / 1000) == 1 ? "sat" : "sats"}
                </p>
              </div>
            )}
            {tx.state != "failed" && type == "outgoing" && (
              <div className="mt-6">
                <p>Fee</p>
                <p className="text-muted-foreground">
                  {new Intl.NumberFormat().format(
                    Math.floor(tx.feesPaid / 1000)
                  )}{" "}
                  {Math.floor(tx.feesPaid / 1000) == 1 ? "sat" : "sats"}
                  {tx.feesPaid > 0 && (
                    <>&nbsp;({((tx.feesPaid / tx.amount) * 100).toFixed(2)}%)</>
                  )}
                </p>
              </div>
            )}
            {tx.description && (
              <div className="mt-6">
                <p>Description</p>
                <p className="text-muted-foreground break-all">
                  {tx.description}
                </p>
              </div>
            )}
            {tx.metadata?.comment && (
              <div className="mt-6">
                <p>Comment</p>
                <p className="text-muted-foreground break-all">
                  {tx.metadata.comment}
                </p>
              </div>
            )}
            {bolt12Offer?.payer_note && (
              <div className="mt-6">
                <p>Payer Note</p>
                <p className="text-muted-foreground break-all">
                  {bolt12Offer.payer_note}
                </p>
              </div>
            )}
            {/* for Alby lightning addresses the content of the zap request is
            automatically extracted and already displayed above as description */}
            {tx.metadata?.nostr && eventId && npub && (
              <div className="mt-6">
                <p>
                  <ExternalLink
                    to={`https://njump.me/${nip19.neventEncode({
                      id: eventId,
                    })}`}
                    className="underline"
                  >
                    Nostr Zap
                  </ExternalLink>{" "}
                  <span className="text-muted-foreground break-all">
                    from {npub}
                  </span>
                </p>
              </div>
            )}
            {tx.state === "failed" && (
              <div className="mt-6">
                <PaymentFailedAlert
                  errorMessage={tx.failureReason}
                  invoice={tx.invoice}
                />
              </div>
            )}
            <div className="mt-4 w-full">
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
                <>
                  {tx.boostagram && <PodcastingInfo boost={tx.boostagram} />}
                  {bolt12Offer && (
                    <div className="mt-6">
                      <p>BOLT-12 Offer Id</p>
                      <div className="flex items-center gap-4">
                        <p className="text-muted-foreground break-all">
                          {bolt12Offer.id}
                        </p>
                        <CopyIcon
                          className="cursor-pointer text-muted-foreground size-4 shrink-0"
                          onClick={() => {
                            copy(bolt12Offer.id as string);
                          }}
                        />
                      </div>
                    </div>
                  )}
                  {tx.preimage && (
                    <div className="mt-6">
                      <p>Preimage</p>
                      <div className="flex items-center gap-4">
                        <p className="text-muted-foreground break-all">
                          {tx.preimage}
                        </p>
                        <CopyIcon
                          className="cursor-pointer text-muted-foreground size-4 shrink-0"
                          onClick={() => {
                            if (tx.preimage) {
                              copy(tx.preimage);
                            }
                          }}
                        />
                      </div>
                    </div>
                  )}
                  <div className="mt-6">
                    <p>Hash</p>
                    <div className="flex items-center gap-4">
                      <p className="text-muted-foreground break-all">
                        {tx.paymentHash}
                      </p>
                      <CopyIcon
                        className="cursor-pointer text-muted-foreground size-4 shrink-0"
                        onClick={() => {
                          copy(tx.paymentHash);
                        }}
                      />
                    </div>
                  </div>
                  <div className="mt-6">
                    <p>Invoice</p>
                    <div className="flex items-center gap-4">
                      <p className="text-muted-foreground break-all">
                        {tx.invoice}
                      </p>
                      <CopyIcon
                        className="cursor-pointer text-muted-foreground size-4 shrink-0"
                        onClick={() => {
                          copy(tx.invoice);
                        }}
                      />
                    </div>
                  </div>
                  {!!tx.failureReason && (
                    <div className="mt-6">
                      <p>Failure Reason</p>
                      <div className="flex items-center gap-4">
                        <p className="text-muted-foreground break-anywhere">
                          {tx.failureReason}
                        </p>
                        <CopyIcon
                          className="cursor-pointer text-muted-foreground size-4 shrink-0"
                          onClick={() => {
                            copy(tx.failureReason);
                          }}
                        />
                      </div>
                    </div>
                  )}
                  {tx.metadata && (
                    <div className="mt-6">
                      <p>Metadata</p>
                      <div className="flex items-center gap-4">
                        <p className="text-muted-foreground break-all">
                          {JSON.stringify(tx.metadata)}
                        </p>
                        <CopyIcon
                          className="cursor-pointer text-muted-foreground size-4 shrink-0"
                          onClick={() => {
                            copy(JSON.stringify(tx.metadata));
                          }}
                        />
                      </div>
                    </div>
                  )}
                </>
              )}
            </div>
          </DialogDescription>
        </DialogHeader>
      </DialogContent>
    </Dialog>
  );
}

export default TransactionItem;
