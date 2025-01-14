import dayjs from "dayjs";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ChevronDown,
  ChevronUp,
  CopyIcon,
  XIcon,
} from "lucide-react";
import { nip19 } from "nostr-tools";
import React from "react";
import { Link } from "react-router-dom";
import AppAvatar from "src/components/AppAvatar";
import ExternalLink from "src/components/ExternalLink";
import PodcastingInfo from "src/components/PodcastingInfo";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import { Transaction } from "src/types";

dayjs.extend(utc);
dayjs.extend(timezone);

type Props = {
  tx: Transaction;
};

function TransactionItem({ tx }: Props) {
  const { data: apps } = useApps();
  const { toast } = useToast();
  const [showDetails, setShowDetails] = React.useState(false);
  const type = tx.type;
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
      : tx.type == "outgoing"
        ? ArrowUpIcon
        : ArrowDownIcon;
  const app =
    tx.appId !== undefined
      ? apps?.find((app) => app.id === tx.appId)
      : undefined;

  const copy = (text: string) => {
    copyToClipboard(text, toast);
  };

  const typeStateIcon = (
    <div className="flex items-center">
      <div
        className={cn(
          "flex justify-center items-center rounded-full w-10 h-10 md:w-14 md:h-14 relative",
          tx.state === "failed"
            ? "bg-red-100 dark:bg-red-950"
            : tx.state === "pending"
              ? "bg-blue-100 dark:bg-blue-900"
              : type === "outgoing"
                ? "bg-orange-100 dark:bg-orange-950"
                : "bg-green-100 dark:bg-emerald-950"
        )}
      >
        <Icon
          strokeWidth={3}
          className={cn(
            "w-6 h-6 md:w-8 md:h-8",
            tx.state === "failed"
              ? "stroke-rose-400 dark:stroke-red-600"
              : tx.state === "pending"
                ? "stroke-blue-500"
                : type === "outgoing"
                  ? "stroke-orange-400 dark:stroke-amber-600"
                  : "stroke-green-400 dark:stroke-emerald-500"
          )}
        />
        {app && (
          <div
            className="absolute -bottom-1 -right-1"
            title={`${typeStateText} via ${app.name === "getalby.com" ? "Alby Account" : app.name}`}
          >
            <AppAvatar
              app={app}
              className="border-none p-0 rounded-full w-[18px] h-[18px] md:w-6 md:h-6 shadow-sm"
            />
          </div>
        )}
      </div>
    </div>
  );

  let from;

  if (tx.metadata?.payer_data?.name) {
    from = "from " + tx.metadata.payer_data.name;
  } else if (tx.metadata?.nostr) {
    const npub = nip19.npubEncode(tx.metadata.nostr.pubkey);
    from = "zap from " + npub.substring(0, 12) + "...";
  }

  const eventId = tx.metadata?.nostr?.tags.find((t) => t[0] === "e")?.[1];

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
            <div>
              <p className="flex items-end truncate">
                <span className="md:text-xl font-semibold">
                  {typeStateText}
                </span>
                {from !== undefined && <>&nbsp;{from}</>}
                <span className="text-xs md:text-base ml-2 truncate text-muted-foreground">
                  {dayjs(tx.settledAt || tx.createdAt).fromNow()}
                </span>
              </p>
            </div>
            <p className="text-sm md:text-base text-muted-foreground break-all w-full truncate">
              {tx.description}
            </p>
          </div>
          <div className="flex ml-auto text-right space-x-3 shrink-0">
            <div className="flex items-center gap-2 md:text-xl">
              <p
                className={cn(
                  "font-semibold",
                  type == "incoming" && "text-green-600 dark:text-emerald-500"
                )}
              >
                {type == "outgoing" ? "-" : "+"}
                {new Intl.NumberFormat().format(
                  Math.floor(tx.amount / 1000)
                )}{" "}
              </p>
              <p className="text-foreground">
                {Math.floor(tx.amount / 1000) == 1 ? "sat" : "sats"}
              </p>

              {/* {!!tx.totalAmountFiat && (
                <p className="text-xs text-muted-foreground">
                  ~{tx.totalAmountFiat}
                </p>
              )} */}
            </div>
          </div>
        </div>
      </DialogTrigger>
      <DialogContent className="slashed-zero">
        <DialogHeader>
          <DialogTitle
            className={cn(tx.state === "pending" && "animate-pulse")}
          >{`${typeStateText} Bitcoin Payment`}</DialogTitle>
          <DialogDescription className="text-start text-foreground">
            <div
              className={cn(
                "flex items-center mt-6",
                tx.state === "pending" && "animate-pulse"
              )}
            >
              {typeStateIcon}
              <div className="ml-4">
                <p className="text-xl md:text-2xl font-semibold sensitive">
                  {new Intl.NumberFormat().format(Math.floor(tx.amount / 1000))}{" "}
                  {Math.floor(tx.amount / 1000) == 1 ? "sat" : "sats"}
                </p>
              </div>
            </div>
            {app && (
              <div className="mt-8">
                <p>App</p>
                <Link to={`/apps/${app.appPubkey}`}>
                  <p className="font-semibold">
                    {app.name === "getalby.com" ? "Alby Account" : app.name}
                  </p>
                </Link>
              </div>
            )}
            <div className="mt-6">
              <p>Date & Time</p>
              <p className="text-muted-foreground">
                {dayjs(tx.settledAt || tx.createdAt)
                  .tz(dayjs.tz.guess())
                  .format("D MMMM YYYY, HH:mm")}
              </p>
            </div>
            {tx.state != "failed" && type == "outgoing" && (
              <div className="mt-6">
                <p>Fee</p>
                <p className="text-muted-foreground">
                  {new Intl.NumberFormat().format(
                    Math.floor(tx.feesPaid / 1000)
                  )}{" "}
                  {Math.floor(tx.feesPaid / 1000) == 1 ? "sat" : "sats"}
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
            {tx.metadata?.nostr && eventId && (
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
                    from {nip19.npubEncode(tx.metadata.nostr.pubkey)}
                  </span>
                </p>
              </div>
            )}
            <div className="mt-4 w-full">
              <div
                className="flex items-center gap-2 cursor-pointer"
                onClick={() => setShowDetails(!showDetails)}
              >
                Details
                {showDetails ? (
                  <ChevronUp className="w-4 h-4" />
                ) : (
                  <ChevronDown className="w-4 h-4" />
                )}
              </div>
              {showDetails && (
                <>
                  {tx.boostagram && <PodcastingInfo boost={tx.boostagram} />}
                  {tx.preimage && (
                    <div className="mt-6">
                      <p>Preimage</p>
                      <div className="flex items-center gap-4">
                        <p className="text-muted-foreground break-all">
                          {tx.preimage}
                        </p>
                        <CopyIcon
                          className="cursor-pointer text-muted-foreground w-4 h-4 flex-shrink-0"
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
                        className="cursor-pointer text-muted-foreground w-4 h-4 flex-shrink-0"
                        onClick={() => {
                          copy(tx.paymentHash);
                        }}
                      />
                    </div>
                  </div>
                  {!!tx.failureReason && (
                    <div className="mt-6">
                      <p>Failure Reason</p>
                      <div className="flex items-center gap-4">
                        <p className="text-muted-foreground break-words">
                          {tx.failureReason}
                        </p>
                        <CopyIcon
                          className="cursor-pointer text-muted-foreground w-4 h-4 flex-shrink-0"
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
                          className="cursor-pointer text-muted-foreground w-4 h-4 flex-shrink-0"
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
