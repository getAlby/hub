import { Drum } from "lucide-react";
import { useMemo, useRef, useState } from "react";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "src/components/ui/pagination";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import { useTransactions } from "src/hooks/useTransactions";
import { cn, generatePageNumbers } from "src/lib/utils";

type TransactionsListProps = {
  appId?: number;
  showReceiveButton?: boolean;
};

function TransactionsList({
  appId,
  showReceiveButton = true,
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
  const totalPages = Math.ceil(totalCount / LIST_TRANSACTIONS_LIMIT);

  const pageNumbers = useMemo(() => {
    return generatePageNumbers(page, totalPages);
  }, [page, totalPages]);

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
    <div ref={transactionListRef} className="transaction-list flex flex-col">
      {!transactions.length ? (
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
          {transactions?.map((tx, i) => {
            return (
              <TransactionItem key={tx.paymentHash + tx.type + i} tx={tx} />
            );
          })}

          {totalPages > 1 && (
            <div className="mt-4 self-center">
              <Pagination>
                <PaginationContent>
                  <PaginationItem
                    className={cn(
                      page === 1 &&
                        "pointer-events-none opacity-30 dark:opacity-20"
                    )}
                  >
                    <PaginationPrevious
                      href="#"
                      onClick={(e) => {
                        e.preventDefault();
                        handlePageChange(page - 1);
                      }}
                    />
                  </PaginationItem>

                  {pageNumbers.map((p, index) =>
                    p === "ellipsis" ? (
                      <PaginationItem key={index}>
                        <PaginationEllipsis className="flex items-center" />
                      </PaginationItem>
                    ) : (
                      <PaginationItem key={index}>
                        <PaginationLink
                          href="#"
                          isActive={p === page}
                          onClick={(e) => {
                            e.preventDefault();
                            handlePageChange(p);
                          }}
                        >
                          {p}
                        </PaginationLink>
                      </PaginationItem>
                    )
                  )}

                  <PaginationItem
                    className={cn(
                      page === totalPages &&
                        "pointer-events-none opacity-30 dark:opacity-20"
                    )}
                  >
                    <PaginationNext
                      href="#"
                      onClick={(e) => {
                        e.preventDefault();
                        handlePageChange(page + 1);
                      }}
                    />
                  </PaginationItem>
                </PaginationContent>
              </Pagination>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default TransactionsList;
