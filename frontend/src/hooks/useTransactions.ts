import useSWR, { SWRConfiguration } from "swr";

import { ListTransactionsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function getTransactionsUrl(appId?: number, limit = 100, page = 1) {
  const offset = (page - 1) * limit;
  let url = `/api/transactions?limit=${limit}&offset=${offset}`;
  if (appId) {
    url += `&appId=${appId}`;
  }

  return url;
}

export function useTransactions(
  appId?: number,
  poll = false,
  limit = 100,
  page = 1
) {
  const url = getTransactionsUrl(appId, limit, page);

  return useSWR<ListTransactionsResponse>(
    url,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
