import useSWR, { SWRConfiguration } from "swr";

import { HealthResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 30000,
};

export function useHealthCheck(poll = true) {
  return useSWR<HealthResponse>(
    "/api/health",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
