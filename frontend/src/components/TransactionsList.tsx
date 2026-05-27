import { ZapIcon } from "lucide-react";
import { useRef, useState } from "react";
import { CustomPagination } from "src/components/CustomPagination";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import { getTransactionsUrl, useTransactions } from "src/hooks/useTransactions";

type TransactionsListProps = {
  appId?: number;
};

function TransactionsList({ appId }: TransactionsListProps) {
  const [page, setPage] = useState(1);
  const transactionListRef = useRef<HTMLDivElement>(null);
  const transactionListKey = getTransactionsUrl(
    appId,
    LIST_TRANSACTIONS_LIMIT,
    page
  );
  const { data: transactionData, isLoading } = useTransactions(
    appId,
    false,
    LIST_TRANSACTIONS_LIMIT,
    page
  );
  const transactions = transactionData?.transactions || [];
  const totalCount = transactionData?.totalCount || 0;

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
      {!transactions.length ? (
        <EmptyState
          icon={ZapIcon}
          title="No lightning payments yet"
          description="Your payments will appear here as you start using your wallet."
          variant="muted"
          showButton={false}
        />
      ) : (
        <>
          {transactions?.map((tx) => {
            return (
              <TransactionItem
                key={tx.id}
                tx={tx}
                transactionListKey={transactionListKey}
              />
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
