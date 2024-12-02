import React from "react";
import { useToast } from "src/components/ui/use-toast";
import { useTransactions } from "src/hooks/useTransactions";
import { Transaction } from "src/types";

export function useNotifyReceivedPayments() {
  const { data: transactions } = useTransactions(undefined, true, 1);
  const [prevTransaction, setPrevTransaction] = React.useState<Transaction>();
  const { toast } = useToast();

  React.useEffect(() => {
    if (transactions && prevTransaction !== transactions[0]) {
      if (prevTransaction && transactions[0].type === "incoming") {
        toast({
          title: "Payment received",
          description: `${new Intl.NumberFormat().format(Math.floor(transactions[0].amount / 1000))} sats`,
        });
      }
      setPrevTransaction(transactions[0]);
    }
  }, [prevTransaction, toast, transactions]);
}
