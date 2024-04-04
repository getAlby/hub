import useSWR from "swr";

import { BalancesResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useBalances() {
  return useSWR<BalancesResponse>("/api/balances", swrFetcher);
}
