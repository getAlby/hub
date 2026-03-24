import useSWR from "swr";

import { BitcoinRate } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useBitcoinRate(poll = false) {
  return useSWR<BitcoinRate>(
    `/api/alby/rates`,
    swrFetcher,
    poll ? { refreshInterval: 3000, refreshWhenHidden: true } : undefined
  );
}
