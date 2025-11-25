import React from "react";
import { toast } from "sonner";
import { BITCOIN_DISPLAY_FORMAT_BIP177 } from "src/constants";
import { useInfo } from "src/hooks/useInfo";
import { useTransactions } from "src/hooks/useTransactions";
import { Transaction } from "src/types";
import { formatBitcoinAmount } from "src/utils/bitcoinFormatting";

export function useNotifyReceivedPayments() {
  const { data: info } = useInfo();
  const { data: transactionsData } = useTransactions(undefined, true, 1);
  const [prevTransaction, setPrevTransaction] = React.useState<Transaction>();

  React.useEffect(() => {
    if (!transactionsData?.transactions?.length || !info) {
      return;
    }
    const latestTx = transactionsData.transactions[0];
    if (latestTx !== prevTransaction) {
      if (prevTransaction && latestTx.type === "incoming") {
        toast("Payment received", {
          description: formatBitcoinAmount(
            latestTx.amount,
            info.bitcoinDisplayFormat || BITCOIN_DISPLAY_FORMAT_BIP177
          ),
        });
      }
      setPrevTransaction(latestTx);
    }
  }, [prevTransaction, transactionsData, info]);
}
