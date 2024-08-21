import dayjs from "dayjs";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ChevronDown,
  ChevronUp,
  CopyIcon,
} from "lucide-react";
import React from "react";
import AppAvatar from "src/components/AppAvatar";
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
  const Icon = tx.type == "outgoing" ? ArrowUpIcon : ArrowDownIcon;
  const app = tx.appId && apps?.find((app) => app.id === tx.appId);

  const copy = (text: string) => {
    copyToClipboard(text);
    toast({ title: "Copied to clipboard." });
  };

  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open) {
          setShowDetails(false);
        }
      }}
    >
      <DialogTrigger className="p-3 mb-4 hover:bg-muted/50 data-[state=selected]:bg-muted cursor-pointer rounded-md slashed-zero transaction sensitive">
        {/* flex wrap is used as a last resort to stop horizontal scrollbar on mobile. */}
        <div className="flex gap-3 flex-wrap">
          <div className="flex items-center">
            {app ? (
              <AppAvatar
                appName={app.name}
                className="border-none p-0 rounded-full w-10 h-10 md:w-14 md:h-14"
              />
            ) : (
              <div
                className={cn(
                  "flex justify-center items-center rounded-full w-10 h-10 md:w-14 md:h-14",
                  type === "outgoing"
                    ? "bg-orange-100 dark:bg-orange-950"
                    : "bg-green-100 dark:bg-emerald-950"
                )}
              >
                <Icon
                  strokeWidth={3}
                  className={cn(
                    "w-6 h-6 md:w-8 md:h-8",
                    type === "outgoing"
                      ? "stroke-orange-400 dark:stroke-amber-600"
                      : "stroke-green-400 dark:stroke-emerald-500"
                  )}
                />
              </div>
            )}
          </div>
          <div className="overflow-hidden mr-3">
            <div className="flex items-center gap-2 truncate">
              <p className="text-lg md:text-xl font-semibold">
                {app ? app.name : type == "incoming" ? "Received" : "Sent"}
              </p>
              <p className="text-sm md:text-base truncate text-muted-foreground">
                {dayjs(tx.settledAt).fromNow()}
              </p>
            </div>
            <p className="text-sm md:text-base text-muted-foreground break-all flex">
              {tx.description}
            </p>
          </div>
          <div className="flex ml-auto text-right space-x-3 shrink-0">
            <div className="flex items-center gap-2 text-xl">
              <p
                className={cn(
                  "font-semibold",
                  type == "incoming" && "text-green-600 dark:text-emerald-500"
                )}
              >
                {type == "outgoing" ? "-" : "+"}
                {new Intl.NumberFormat(undefined, {}).format(
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
          <DialogTitle>{`${type == "outgoing" ? "Sent" : "Received"} Bitcoin`}</DialogTitle>
          <DialogDescription className="text-start text-foreground">
            <div className="flex items-center mt-6">
              <div
                className={cn(
                  "flex justify-center items-center rounded-full w-10 h-10 md:w-14 md:h-14",
                  type === "outgoing"
                    ? "bg-orange-100 dark:bg-orange-950"
                    : "bg-green-100 dark:bg-emerald-950"
                )}
              >
                <Icon
                  strokeWidth={3}
                  className={cn(
                    "w-6 h-6 md:w-8 md:h-8",
                    type === "outgoing"
                      ? "stroke-orange-400 dark:stroke-amber-600"
                      : "stroke-green-400 dark:stroke-emerald-500"
                  )}
                />
              </div>
              <div className="ml-4">
                <p className="text-xl md:text-2xl font-semibold">
                  {new Intl.NumberFormat(undefined, {}).format(
                    Math.floor(tx.amount / 1000)
                  )}{" "}
                  {Math.floor(tx.amount / 1000) == 1 ? "sat" : "sats"}
                </p>
                {/* <p className="text-sm md:text-base text-muted-foreground">
                Fiat Amount
              </p> */}
              </div>
            </div>
            <div className="mt-8">
              <p>Date & Time</p>
              <p className="text-muted-foreground">
                {dayjs(tx.settledAt)
                  .tz(dayjs.tz.guess())
                  .format("D MMMM YYYY, HH:mm")}
              </p>
            </div>
            {type == "outgoing" && (
              <div className="mt-6">
                <p>Fee</p>
                <p className="text-muted-foreground">
                  {new Intl.NumberFormat(undefined, {}).format(
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
                  <div className="mt-6">
                    <p>Preimage</p>
                    <div className="flex items-center gap-4">
                      <p className="text-muted-foreground break-all">
                        {tx.preimage}
                      </p>
                      <CopyIcon
                        className="cursor-pointer text-muted-foreground w-6 h-6"
                        onClick={() => {
                          if (tx.preimage) {
                            copy(tx.preimage);
                          }
                        }}
                      />
                    </div>
                  </div>
                  <div className="mt-6">
                    <p>Hash</p>
                    <div className="flex items-center gap-4">
                      <p className="text-muted-foreground break-all">
                        {tx.paymentHash}
                      </p>
                      <CopyIcon
                        className="cursor-pointer text-muted-foreground w-6 h-6"
                        onClick={() => {
                          copy(tx.paymentHash);
                        }}
                      />
                    </div>
                  </div>
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
