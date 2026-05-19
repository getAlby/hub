import useSWR, { SWRConfiguration } from "swr";

import { Channel } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useChannels(poll = false, enabled = true) {
  return useSWR<Channel[]>(
    enabled ? "/api/channels" : undefined,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
