import useSWR, { SWRConfiguration } from "swr";

import { Transaction } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useTransactions(poll = false, limit = 20, page = 1) {
  const offset = (page - 1) * limit;
  return useSWR<Transaction[]>(
    `/api/transactions?limit=${limit}&offset=${offset}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
