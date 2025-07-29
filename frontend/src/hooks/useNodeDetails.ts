import useSWR, { SWRConfiguration } from "swr";

import { MempoolNode } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useNodeDetails(nodePubkey: string, poll?: boolean | number) {
  const config: SWRConfiguration = {
    dedupingInterval: 600000, // 10 minutes
  };
  if (typeof poll === "number") {
    config.refreshInterval = poll;
  } else if (poll) {
    config.refreshInterval = 10000;
  }

  return useSWR<MempoolNode>(
    `/api/mempool?endpoint=/v1/lightning/nodes/${nodePubkey}`,
    swrFetcher,
    config
  );
}
