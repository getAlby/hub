import useSWR, { SWRConfiguration } from "swr";

import { BalancesResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  // SWR also revalidates on window focus, so a longer interval keeps the
  // balance fresh without flooding the backend from long-lived open tabs.
  refreshInterval: 10000,
};

export function useBalances(poll = false) {
  return useSWR<BalancesResponse>(
    "/api/balances",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
