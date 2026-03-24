import useSWR, { SWRConfiguration } from "swr";

import { BalancesResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
  refreshWhenHidden: true,
};

export function useBalances(poll = false) {
  return useSWR<BalancesResponse>(
    "/api/balances",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
