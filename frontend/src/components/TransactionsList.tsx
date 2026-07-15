import { LucideIcon, ZapIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { CustomPagination } from "src/components/CustomPagination";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import {
  getTransactionsUrl,
  type TransactionFilters,
  useTransactions,
} from "src/hooks/useTransactions";

type TransactionsListProps = {
  appId?: number;
  filters?: TransactionFilters;
  emptyIcon?: LucideIcon;
  emptyTitle?: string;
  emptyDescription?: string;
  emptyVariant?: "dashed" | "muted" | "none";
};

function TransactionsList({
  appId,
  filters,
  emptyIcon = ZapIcon,
  emptyTitle = "No lightning payments yet",
  emptyDescription = "Your payments will appear here as you start using your wallet.",
  emptyVariant,
}: TransactionsListProps) {
  const [page, setPage] = useState(1);
  const transactionListRef = useRef<HTMLDivElement>(null);
  const transactionListKey = getTransactionsUrl(
    appId,
    LIST_TRANSACTIONS_LIMIT,
    page,
    filters
  );
  const { data: transactionData, isLoading } = useTransactions(
    appId,
    false,
    LIST_TRANSACTIONS_LIMIT,
    page,
    filters
  );
  const transactions = transactionData?.transactions || [];
  const totalCount = transactionData?.totalCount || 0;
  const hasActiveFilters =
    !!filters &&
    ((filters.minAmountSat ?? 0) > 0 || filters.showFailed === false);

  useEffect(() => {
    setPage(1);
  }, [appId, filters?.minAmountSat, filters?.showFailed]);

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
          icon={emptyIcon}
          title={hasActiveFilters ? "No matching payments" : emptyTitle}
          description={
            hasActiveFilters
              ? "Try changing your filters to see more payments."
              : emptyDescription
          }
          variant={emptyVariant}
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
