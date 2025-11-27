import { DrumIcon } from "lucide-react";
import { useRef, useState } from "react";
import { CustomPagination } from "src/components/CustomPagination";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { Button } from "src/components/ui/button";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import { useTransactions } from "src/hooks/useTransactions";

type TransactionsListProps = {
  appId?: number;
  showReceiveButton?: boolean;
  minAmountSats?: number;
  onFilterChange?: (value: number | undefined) => void;
};

function TransactionsList({
  appId,
  showReceiveButton = true,
  minAmountSats,
  onFilterChange,
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
      {!transactions.length ? (
        <EmptyState
          icon={DrumIcon}
          title="No transactions yet"
          description="Your most recent incoming and outgoing payments will show up here."
          buttonText="Receive Your First Payment"
          buttonLink="/wallet/receive"
          showButton={showReceiveButton}
          showBorder={false}
        />
      ) : minAmountSats && !filteredTransactions.length ? (
        <div className="flex flex-1 items-center justify-center rounded-lg p-8">
          <div className="flex flex-col items-center gap-1 text-center max-w-sm">
            <DrumIcon className="w-10 h-10 text-muted-foreground" />
            <h3 className="mt-4 text-lg font-semibold">
              No transactions found
            </h3>
            <p className="text-sm text-muted-foreground">
              You don't have any transactions of{" "}
              {minAmountSats?.toLocaleString()} sats or more yet. Try selecting
              a lower amount to see your transaction history.
            </p>
            {onFilterChange && (
              <Button
                onClick={() => onFilterChange(undefined)}
                className="mt-4"
              >
                Show All Transactions
              </Button>
            )}
          </div>
        </div>
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
