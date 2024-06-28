import useSWR, { SWRConfiguration } from "swr";

import { Transaction } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useTransaction(paymentHash: string, poll = false) {
  return useSWR<Transaction>(
    paymentHash && `/api/transactions/${paymentHash}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
