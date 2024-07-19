import { Drum } from "lucide-react";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { useTransactions } from "src/hooks/useTransactions";

function TransactionsList() {
  const { data: transactions, isLoading } = useTransactions();

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="transaction-list">
      {!transactions?.length ? (
        <EmptyState
          icon={Drum}
          title="No transactions yet"
          description="Your most recent incoming and outgoing payments will show up here."
          buttonText="Receive Your First Payment"
          buttonLink="/wallet/receive"
        />
      ) : (
        <>
          {transactions?.map((tx) => {
            return <TransactionItem key={tx.payment_hash + tx.type} tx={tx} />;
          })}
        </>
      )}
    </div>
  );
}

export default TransactionsList;
