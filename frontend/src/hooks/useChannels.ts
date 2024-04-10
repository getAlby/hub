import useSWR, { SWRConfiguration } from "swr";

import { Channel } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useChannels(poll = false) {
  return useSWR<Channel[]>(
    "/api/channels",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
