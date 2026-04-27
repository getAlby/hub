import { toast } from "sonner";
import { LIST_TRANSACTIONS_LIMIT } from "src/constants";
import { ListTransactionsResponse, Transaction } from "src/types";
import { request } from "src/utils/request";

const LABEL_COLUMN_PREFIX = "label_";

const escapeCsvCell = (raw: string) => {
  // prevent CSV injection: prefix cells starting with =, +, -, @ with a quote
  const safe = /^[\t\r ]*[=+\-@]/.test(raw) ? `'${raw}` : raw;
  return `"${safe.replaceAll('"', '""')}"`;
};

const stringifyCell = (value: unknown) => {
  if (value === undefined || value === null) {
    return "";
  }
  return typeof value === "object" ? JSON.stringify(value) : String(value);
};

export const convertToCSV = (transactions: Transaction[]) => {
  if (!transactions.length) {
    return "";
  }

  const baseHeaders = Object.keys(transactions[0]);

  // Collect the union of all user_label keys across the export so each one
  // becomes its own column. Sorted for stable output between exports.
  const labelKeys = Array.from(
    new Set(
      transactions.flatMap((tx) =>
        tx.metadata?.user_label ? Object.keys(tx.metadata.user_label) : []
      )
    )
  ).sort();
  const labelHeaders = labelKeys.map((key) => `${LABEL_COLUMN_PREFIX}${key}`);

  const csvHeaders = [...baseHeaders, ...labelHeaders]
    .map(escapeCsvCell)
    .join(",");

  const csvRows = transactions.map((tx) => {
    const baseCells = baseHeaders.map((header) =>
      escapeCsvCell(stringifyCell(tx[header as keyof typeof tx]))
    );
    const labelCells = labelKeys.map((key) =>
      escapeCsvCell(stringifyCell(tx.metadata?.user_label?.[key]))
    );
    return [...baseCells, ...labelCells].join(",");
  });

  return [csvHeaders, ...csvRows].join("\n");
};

export const handleExportTransactions = async (appId?: number) => {
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
