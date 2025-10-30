import React from "react";
import { toast } from "sonner";
import { useTransactions } from "src/hooks/useTransactions";
import { Transaction } from "src/types";
import { formatBitcoinAmount } from "src/utils/bitcoinFormatting";

export function useNotifyReceivedPayments() {
  const { data: transactionsData } = useTransactions(undefined, true, 1);
  const [prevTransaction, setPrevTransaction] = React.useState<Transaction>();

  React.useEffect(() => {
    if (!transactionsData?.transactions?.length) {
      return;
    }
    const latestTx = transactionsData.transactions[0];
    if (latestTx !== prevTransaction) {
      if (prevTransaction && latestTx.type === "incoming") {
        toast("Payment received", {
          description: formatBitcoinAmount(latestTx.amount),
        });
      }
      setPrevTransaction(latestTx);
    }
  }, [prevTransaction, transactionsData]);
}
