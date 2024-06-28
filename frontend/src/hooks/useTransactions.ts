import useSWR, { SWRConfiguration } from "swr";

import { Transaction } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useTransactions(poll = false) {
  return useSWR<Transaction[]>(
    "/api/transactions",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
