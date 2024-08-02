import useSWR from "swr";

import { AlbyBalance } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAlbyBalance() {
  return useSWR<AlbyBalance>("/api/alby/balance", swrFetcher, {
    dedupingInterval: 5 * 60 * 1000, // 5 minutes
  });
}
