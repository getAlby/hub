import useSWR, { SWRConfiguration } from "swr";

import { BalancesResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 10000,
};

export function useBalances(poll = false) {
  return useSWR<BalancesResponse>(
    "/api/balances",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
