import { DrumIcon } from "lucide-react";
import { useRef, useState } from "react";
import { CustomPagination } from "src/components/CustomPagination";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import { useTransactions } from "src/hooks/useTransactions";

type TransactionsListProps = {
  appId?: number;
  showReceiveButton?: boolean;
  minAmountSats?: number;
};

function TransactionsList({
  appId,
  showReceiveButton = true,
  minAmountSats,
}: TransactionsListProps) {
  const [page, setPage] = useState(1);
  const transactionListRef = useRef<HTMLDivElement>(null);
  const { data: transactionData, isLoading } = useTransactions(
    appId,
    false,
    LIST_TRANSACTIONS_LIMIT,
    page
  );
  const transactions = transactionData?.transactions || [];
  const totalCount = transactionData?.totalCount || 0;

  const filteredTransactions = minAmountSats
    ? transactions.filter((tx) => Math.floor(tx.amount / 1000) >= minAmountSats)
    : transactions;

  const handlePageChange = (page: number) => {
    setPage(page);
    transactionListRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  };

  if (isLoading || !transactionData) {
    return <Loading />;
  }

  return (
    <div ref={transactionListRef} className="flex flex-col flex-1">
      {!filteredTransactions.length ? (
        <EmptyState
          icon={DrumIcon}
          title="No transactions yet"
          description="Your most recent incoming and outgoing payments will show up here."
          buttonText="Receive Your First Payment"
          buttonLink="/wallet/receive"
          showButton={showReceiveButton}
          showBorder={false}
        />
      ) : (
        <>
          {filteredTransactions?.map((tx, i) => {
            return (
              <TransactionItem key={tx.paymentHash + tx.type + i} tx={tx} />
            );
          })}
          <CustomPagination
            limit={LIST_TRANSACTIONS_LIMIT}
            totalCount={totalCount}
            page={page}
            handlePageChange={handlePageChange}
          />
        </>
      )}
    </div>
  );
}

export default TransactionsList;
