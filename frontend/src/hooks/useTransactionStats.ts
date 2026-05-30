import useSWR from "swr";

import { GetTransactionStatsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useTransactionStats() {
  return useSWR<GetTransactionStatsResponse>(
    "/api/transactions/stats",
    swrFetcher
  );
}
