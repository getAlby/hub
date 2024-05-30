import useSWR, { SWRConfiguration } from "swr";

import { Peer } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function usePeers(poll = false) {
  return useSWR<Peer[]>(
    "/api/peers",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
