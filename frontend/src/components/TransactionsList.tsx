import { Drum } from "lucide-react";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { useTransactions } from "src/hooks/useTransactions";

type TransactionsListProps = {
  appId?: number;
  showReceiveButton?: boolean;
};

function TransactionsList({
  appId,
  showReceiveButton = true,
}: TransactionsListProps) {
  const { data: transactions, isLoading } = useTransactions(appId);

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="transaction-list flex flex-col">
      {!transactions?.length ? (
        <EmptyState
          icon={Drum}
          title="No transactions yet"
          description="Your most recent incoming and outgoing payments will show up here."
          buttonText="Receive Your First Payment"
          buttonLink="/wallet/receive"
          showButton={showReceiveButton}
        />
      ) : (
        <>
          {transactions?.map((tx) => {
            return <TransactionItem key={tx.paymentHash + tx.type} tx={tx} />;
          })}
        </>
      )}
    </div>
  );
}

export default TransactionsList;
