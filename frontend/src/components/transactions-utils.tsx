import { toast } from "sonner";
import { ListTransactionsResponse, Transaction } from "src/types";
import { request } from "src/utils/request";

const LABEL_COLUMN_PREFIX = "label_";

// Use a large page size to keep the number of round-trips low when
// exporting wallets that have thousands of transactions.
const EXPORT_TRANSACTIONS_PAGE_SIZE = 1000;

// based on https://stackoverflow.com/a/68146412
const escapeCsvCell = (raw: string) => {
  const safe = /^[\t\r ]*[=+\-@]/.test(raw) ? `'${raw}` : raw;
  return `"${safe.replaceAll('"', '""')}"`;
};

export const convertToCSV = (transactions: Transaction[]) => {
  if (!transactions.length) {
    return "";
  }

  // Get headers from all transactions including the user_labels in metadata
  const headers = Object.keys(transactions[0]);
  const userLabelKeys = Array.from(
    new Set(
      transactions.flatMap((tx) =>
        tx.metadata?.user_labels ? Object.keys(tx.metadata.user_labels) : []
      )
    )
  ).sort();
  const userLabelHeaders = userLabelKeys.map(
    (key) => `${LABEL_COLUMN_PREFIX}${key}`
  );

  const csvHeaders = [...headers, ...userLabelHeaders].join(",");

  // Convert each transaction to CSV row
  const csvRows = transactions.map((tx) => {
    return [
      ...headers.map((header) => {
        const value = tx[header as keyof typeof tx];
        if (value === undefined || value === null) {
          return "";
        }
        const stringValue =
          typeof value === "object" ? JSON.stringify(value) : String(value);
        return escapeCsvCell(stringValue);
      }),
      ...userLabelKeys.map((key) => {
        const value = tx.metadata?.user_labels?.[key];
        if (value === undefined || value === null) {
          return "";
        }
        return escapeCsvCell(value);
      }),
    ].join(",");
  });

  return [csvHeaders, ...csvRows].join("\n");
};

export const handleExportTransactions = async (appId?: number) => {
  const toastId = toast.loading("Exporting transactions…");
  try {
    // Fetch all transactions by paginating through all pages
    let allTransactions: Transaction[] = [];
    let offset = 0;

    while (true) {
      let url = `/api/transactions?limit=${EXPORT_TRANSACTIONS_PAGE_SIZE}&offset=${offset}`;
      if (appId) {
        url += `&appId=${appId}`;
      }

      const data = await request<ListTransactionsResponse>(url);

      if (!data) {
        throw new Error("no list transactions response");
      }

      allTransactions = [...allTransactions, ...data.transactions];
      toast.loading(
        `Exporting ${allTransactions.length.toLocaleString()} transactions…`,
        { id: toastId }
      );

      if (data.transactions.length < EXPORT_TRANSACTIONS_PAGE_SIZE) {
        break;
      }
      offset += EXPORT_TRANSACTIONS_PAGE_SIZE;
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
    toast.success("Transactions saved to your downloads folder", {
      id: toastId,
    });
  } catch (error) {
    console.error("Error downloading transactions:", error);
    toast.error("Failed to export transactions", { id: toastId });
  }
};
