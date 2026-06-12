import useSWR, { SWRConfiguration } from "swr";

import { ListTransactionsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 10000,
};

export function getTransactionsUrl(
  appId?: number,
  limit = 100,
  page = 1,
  searchTerm = ""
) {
  const offset = (page - 1) * limit;
  let url = `/api/transactions?limit=${limit}&offset=${offset}`;
  if (appId) {
    url += `&appId=${appId}`;
  }
  if (searchTerm) {
    url += `&q=${encodeURIComponent(searchTerm)}`;
  }

  return url;
}

export function useTransactions(
  appId?: number,
  poll = false,
  limit = 100,
  page = 1,
  searchTerm = ""
) {
  const url = getTransactionsUrl(appId, limit, page, searchTerm);

  return useSWR<ListTransactionsResponse>(url, swrFetcher, {
    ...(poll ? pollConfiguration : {}),
    keepPreviousData: true,
  });
}
