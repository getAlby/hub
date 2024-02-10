import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { NodeConnectionInfo } from "src/types";

export function useNodeConnectionInfo() {
  return useSWR<NodeConnectionInfo>("/api/node", swrFetcher);
}
