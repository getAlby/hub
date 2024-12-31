import useSWR from "swr";

import { BitcoinRate } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useBitcoinRate(currency: string) {
  return useSWR<BitcoinRate>(
    `/api/alby/btc_rate?currency=${currency}`,
    swrFetcher,
    {
      dedupingInterval: 5 * 60 * 1000, // 5 minutes
    }
  );
}
