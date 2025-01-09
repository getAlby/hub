import useSWR from "swr";

import { BitcoinRate } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useBitcoinRate() {
  return useSWR<BitcoinRate>(`/api/alby/rates`, swrFetcher, {
    dedupingInterval: 5 * 60 * 1000, // 5 minutes
  });
}
