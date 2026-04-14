import useSWR from "swr";

import { BitcoinRate } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useBitcoinRate(currency?: string) {
  const shouldFetch = currency && currency !== "SATS";

  return useSWR<BitcoinRate>(
    shouldFetch ? `/api/alby/rates/${encodeURIComponent(currency)}` : null,
    swrFetcher
  );
}
