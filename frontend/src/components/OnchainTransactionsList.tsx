import { LinkIcon } from "lucide-react";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import OnchainTransactionItem from "src/components/OnchainTransactionItem";
import { useInfo } from "src/hooks/useInfo";
import { useOnchainTransactions } from "src/hooks/useOnchainTransactions";

export function OnchainTransactionsList() {
  const { data: info } = useInfo();
  // TODO: add pagination
  const { data: transactions, isLoading } = useOnchainTransactions();

  if (isLoading || !transactions) {
    return <Loading />;
  }

  if (transactions.length === 0) {
    return (
      <div className="flex w-full flex-1 flex-col">
        <EmptyState
          icon={LinkIcon}
          title="No on-chain transactions yet"
          description="Your most recent incoming and outgoing on-chain transactions will show up here."
          buttonText="Receive to On-chain Balance"
          buttonLink="/wallet/receive?type=onchain"
          showBorder={false}
        />
      </div>
    );
  }

  return (
    <div className="flex w-full flex-1 flex-col space-y-4">
      {transactions.map((tx) => (
        <OnchainTransactionItem
          key={tx.txId}
          tx={tx}
          mempoolUrl={info?.mempoolUrl}
        />
      ))}
    </div>
  );
}
