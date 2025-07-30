import useSWR, { SWRConfiguration } from "swr";

import { MempoolNode } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useNodeDetails(nodePubkey: string | undefined) {
  const config: SWRConfiguration = {
    dedupingInterval: 600000, // 10 minutes
  };

  return useSWR<MempoolNode>(
    nodePubkey && `/api/mempool?endpoint=/v1/lightning/nodes/${nodePubkey}`,
    swrFetcher,
    config
  );
}
