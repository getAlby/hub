import useSWR from "swr";

import { AlbyInfo } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAlbyInfo() {
  return useSWR<AlbyInfo>("/api/alby/info", swrFetcher, {
    dedupingInterval: 5 * 60 * 1000, // 5 minutes
  });
}
