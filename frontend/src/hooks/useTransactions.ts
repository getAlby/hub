import useSWR, { SWRConfiguration } from "swr";

import { ListTransactionsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useTransactions(
  appId?: number,
  poll = false,
  limit = 100,
  page = 1
) {
  const offset = (page - 1) * limit;
  return useSWR<ListTransactionsResponse>(
    `/api/transactions?limit=${limit}&offset=${offset}&appId=${appId}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
