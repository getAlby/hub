import useSWR, { SWRConfiguration } from "swr";

import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 10000,
};

export function useMempoolApi<T>(endpoint: string | undefined, poll = false) {
  return useSWR<T>(
    endpoint && `/api/mempool?endpoint=${endpoint}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
