import useSWR, { SWRConfiguration } from "swr";

import { ListTransactionsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 10000,
};

export type TransactionFilters = {
  minAmountSat?: number;
  hideFailed?: boolean;
};

export const defaultTransactionFilters: TransactionFilters = {};

export function getTransactionsUrl(
  appId?: number,
  limit = 100,
  page = 1,
  filters?: TransactionFilters
) {
  const offset = (page - 1) * limit;
  const searchParams = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });

  if (appId) {
    searchParams.set("appId", String(appId));
  }
  if (filters?.minAmountSat && filters.minAmountSat > 0) {
    searchParams.set("minAmountSat", String(filters.minAmountSat));
  }
  if (filters?.hideFailed) {
    searchParams.set("hideFailed", "true");
  }

  const url = `/api/transactions?${searchParams.toString()}`;

  return url;
}

export function useTransactions(
  appId?: number,
  poll = false,
  limit = 100,
  page = 1,
  filters?: TransactionFilters
) {
  const url = getTransactionsUrl(appId, limit, page, filters);

  return useSWR<ListTransactionsResponse>(
    url,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
