import useSWR, { SWRConfiguration } from "swr";

import { NodeStatus } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useNodeStatus(enabled = true) {
  return useSWR<NodeStatus | null>(
    enabled ? "/api/node/status" : null,
    swrFetcher,
    pollConfiguration
  );
}
