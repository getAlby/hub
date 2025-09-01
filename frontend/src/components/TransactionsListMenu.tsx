import { DownloadIcon, MoreHorizontalIcon } from "lucide-react";
import { toast } from "sonner";
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
import { ListTransactionsResponse, Transaction } from "src/types";
import { request } from "src/utils/request";

const convertToCSV = (transactions: Transaction[]) => {
  if (!transactions.length) {
    return "";
  }

  // Get headers from all transactions
  const headers = Object.keys(transactions[0]);
  const csvHeaders = headers.join(",");

  // Convert each transaction to CSV row
  const csvRows = transactions.map((tx) => {
    return headers
      .map((header) => {
        const value = tx[header as keyof typeof tx];
        if (value === undefined || value === null) {
          return "";
        }
        // based on https://stackoverflow.com/a/68146412
        return `"${
          value
            .toString() // convert every value to String
            .replaceAll('"', '""') // escape double quotes
        }"`; // quote it
      })
      .join(",");
  });

  return [csvHeaders, ...csvRows].join("\n");
};

const handleExportTransactions = async (appId?: number) => {
  try {
    // Fetch all transactions by paginating through all pages
    let allTransactions: Transaction[] = [];
    let offset = 0;

    while (true) {
      let url = `/api/transactions?limit=${LIST_TRANSACTIONS_LIMIT}&offset=${offset}`;
      if (appId) {
        url += `&appId=${appId}`;
      }

      const data = await request<ListTransactionsResponse>(url);

      if (!data) {
        throw new Error("no list transactions response");
      }

      allTransactions = [...allTransactions, ...data.transactions];

      if (data.transactions.length < LIST_TRANSACTIONS_LIMIT) {
        break;
      }
      offset += LIST_TRANSACTIONS_LIMIT;
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
    toast("Transactions saved to your downloads folder");
  } catch (error) {
    console.error("Error downloading transactions:", error);
    toast.error("Failed to export transactions");
  }
};

export const TransactionsListMenu = ({ appId }: { appId?: number }) => {
  const { data: albyMe } = useAlbyMe();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger>
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
            onClick={() => handleExportTransactions(appId)}
          >
            <DownloadIcon className="h-4 w-4" />
            Export Transactions
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
