import useSWR, { SWRConfiguration } from "swr";

import { InfoResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useInfo(poll = false) {
  return useSWR<InfoResponse>(
    "/api/info",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
