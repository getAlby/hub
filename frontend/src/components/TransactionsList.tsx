import { LucideIcon, SearchIcon, XIcon, ZapIcon } from "lucide-react";
import { useRef, useState } from "react";
import { useSearchParams } from "react-router";
import { CustomPagination } from "src/components/CustomPagination";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { Badge } from "src/components/ui/badge";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import { getTransactionsUrl, useTransactions } from "src/hooks/useTransactions";

type TransactionsListProps = {
  appId?: number;
  emptyIcon?: LucideIcon;
  emptyTitle?: string;
  emptyDescription?: string;
  emptyVariant?: "dashed" | "muted" | "none";
};

function TransactionsList({
  appId,
  emptyIcon = ZapIcon,
  emptyTitle = "No lightning payments yet",
  emptyDescription = "Your payments will appear here as you start using your wallet.",
  emptyVariant,
}: TransactionsListProps) {
  const [page, setPage] = useState(1);
  const [searchParams, setSearchParams] = useSearchParams();
  const searchTerm = searchParams.get("q") ?? "";
  const transactionListRef = useRef<HTMLDivElement>(null);
  const transactionListKey = getTransactionsUrl(
    appId,
    LIST_TRANSACTIONS_LIMIT,
    page,
    searchTerm
  );
  const { data: transactionData } = useTransactions(
    appId,
    false,
    LIST_TRANSACTIONS_LIMIT,
    page,
    searchTerm
  );
  const transactions = transactionData?.transactions || [];
  const totalCount = transactionData?.totalCount || 0;

  const clearSearch = () => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete("q");
        return next;
      },
      { replace: true }
    );
    setPage(1);
  };

  const handlePageChange = (page: number) => {
    setPage(page);
    transactionListRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  };

  if (!transactionData) {
    return <Loading />;
  }

  return (
    <div ref={transactionListRef} className="flex flex-col flex-1">
      {!!searchTerm && (
        <div className="flex justify-end mb-2">
          <Badge variant="secondary" className="max-w-56 gap-1">
            <SearchIcon className="size-3 shrink-0" />
            <span className="truncate sensitive">{searchTerm}</span>
            <button
              onClick={clearSearch}
              className="cursor-pointer shrink-0"
              aria-label="Clear search"
            >
              <XIcon className="size-3" />
            </button>
          </Badge>
        </div>
      )}
      {!transactions.length ? (
        searchTerm ? (
          <EmptyState
            icon={SearchIcon}
            title="No transactions found"
            description="Try a different search term."
            variant={emptyVariant}
          />
        ) : (
          <EmptyState
            icon={emptyIcon}
            title={emptyTitle}
            description={emptyDescription}
            variant={emptyVariant}
          />
        )
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
