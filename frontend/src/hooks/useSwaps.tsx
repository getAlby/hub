import useSWR, { SWRConfiguration } from "swr";

import { SwapsSettingsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 5 * 60 * 1000, // 5 minutes
};

export function useSwaps(poll = true) {
  return useSWR<SwapsSettingsResponse[]>(
    "/api/wallet/autoswap/out",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
