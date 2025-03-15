import React from "react";
import { useToast } from "src/components/ui/use-toast";
import { useTransactions } from "src/hooks/useTransactions";
import { Transaction } from "src/types";

export function useNotifyReceivedPayments() {
  const { data: transactionsData } = useTransactions(undefined, true, 1);
  const [prevTransaction, setPrevTransaction] = React.useState<Transaction>();
  const { toast } = useToast();

  React.useEffect(() => {
    if (!transactionsData?.transactions?.length) {
      return;
    }
    const latestTx = transactionsData.transactions[0];
    if (latestTx !== prevTransaction) {
      if (prevTransaction && latestTx.type === "incoming") {
        toast({
          title: "Payment received",
          description: `${new Intl.NumberFormat().format(Math.floor(latestTx.amount / 1000))} sats`,
        });
      }
      setPrevTransaction(latestTx);
    }
  }, [prevTransaction, toast, transactionsData]);
}
