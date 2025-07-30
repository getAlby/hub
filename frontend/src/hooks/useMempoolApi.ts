import useSWR, { SWRConfiguration } from "swr";

import { swrFetcher } from "src/utils/swr";

export function useMempoolApi<T>(
  endpoint: string | undefined,
  poll?: boolean | number
) {
  const config: SWRConfiguration | undefined =
    typeof poll === "number"
      ? { refreshInterval: poll }
      : poll
        ? { refreshInterval: 10000 }
        : undefined;

  return useSWR<T>(
    endpoint && `/api/mempool?endpoint=${endpoint}`,
    swrFetcher,
    config
  );
}
