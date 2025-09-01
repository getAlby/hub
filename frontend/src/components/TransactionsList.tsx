import { DownloadIcon, DrumIcon, MoreHorizontalIcon } from "lucide-react";
import { useRef, useState } from "react";
import { CustomPagination } from "src/components/CustomPagination";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import TransactionItem from "src/components/TransactionItem";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useTransactions } from "src/hooks/useTransactions";
import { ListTransactionsResponse, Transaction } from "src/types";
import { request } from "src/utils/request";

const convertToCSV = (transactions: Transaction[]) => {
  if (!transactions.length) {
    return "";
  }

  // Get headers from the first transaction
  const headers = Object.keys(transactions[0]);
  const csvHeaders = headers.join(",");

  // Convert each transaction to CSV row
  const csvRows = transactions.map((tx) => {
    return headers
      .map((header) => {
        const value = (tx as Record<string, unknown>)[header];
        // Escape commas and quotes in values
        if (
          typeof value === "string" &&
          (value.includes(",") || value.includes('"'))
        ) {
          return `"${value.replace(/"/g, '""')}"`;
        }
        return value;
      })
      .join(",");
  });

  return [csvHeaders, ...csvRows].join("\n");
};

const handleDownloadTransactions = async (appId?: number) => {
  try {
    // Fetch all transactions by paginating through all pages
    let allTransactions: Transaction[] = [];
    let offset = 0;
    let hasMoreTransactions = true;

    while (hasMoreTransactions) {
      let url = `/api/transactions?limit=${LIST_TRANSACTIONS_LIMIT}&offset=${offset}`;
      if (appId) {
        url += `&appId=${appId}`;
      }

      const data = (await request(url)) as ListTransactionsResponse;

      if (data.transactions && data.transactions.length > 0) {
        allTransactions = [...allTransactions, ...data.transactions];
        offset += LIST_TRANSACTIONS_LIMIT;
      } else {
        hasMoreTransactions = false;
      }
    }

    // Convert to CSV and create download
    const csvString = convertToCSV(allTransactions);
    const blob = new Blob([csvString], { type: "text/csv" });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    const filename = appId
      ? `transactions_app_${appId}.csv`
      : `transactions_all.csv`;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error("Error downloading transactions:", error);
  }
};

export const TransactionsExportMenu = ({ appId }: { appId?: number }) => {
  const { data: albyMe } = useAlbyMe();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="icon" variant="ghost">
          <MoreHorizontalIcon className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {!albyMe?.subscription.plan_code ? (
          <UpgradeDialog>
            <div className="cursor-pointer">
              <DropdownMenuItem className="w-full pointer-events-none">
                <div className="w-full flex items-center">
                  <DownloadIcon className="h-4 w-4 mr-2" />
                  Export Transactions
                </div>
              </DropdownMenuItem>
            </div>
          </UpgradeDialog>
        ) : (
          <DropdownMenuItem
            className="flex flex-row items-center gap-2 cursor-pointer"
            onClick={() => handleDownloadTransactions(appId)}
          >
            <DownloadIcon className="h-4 w-4" />
            Export Transactions
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

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
      ) : (
        <>
          {!appId && transactions.length > 0 && (
            <div className="mb-4 flex justify-end">
              <TransactionsExportMenu />
            </div>
          )}
          {transactions?.map((tx, i) => {
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
