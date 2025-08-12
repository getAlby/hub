import dayjs from "dayjs";
import { InfoIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LinkButton } from "src/components/ui/button";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useTransactions } from "src/hooks/useTransactions";
import { cn } from "src/lib/utils";

export function SendAlert({ className }: { className?: string }) {
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { data: transactionData } = useTransactions();

  if (!balances || !channels || !transactionData) {
    return null;
  }

  if (
    !transactionData.transactions.some(
      (tx) =>
        tx.state === "pending" &&
        dayjs().diff(dayjs(tx.createdAt)) <
          1000 * 60 * 60 * 24 /* payment pending in last 24h */
      /* TODO: remove diff check when expired transactions are marked as failed */
    )
  ) {
    return null;
  }

  return (
    <Alert className={cn("md:max-w-lg", className)}>
      <InfoIcon className="h-4 w-4" />
      <AlertTitle>Pending Payment</AlertTitle>
      <AlertDescription>
        <div className="mb-2">
          You have one or more payments that have not settled.
        </div>
        <LinkButton to={"/wallet"} size={"sm"}>
          View Payments
        </LinkButton>
      </AlertDescription>
    </Alert>
  );
}
