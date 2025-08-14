import useSWR, { SWRConfiguration } from "swr";

import { AutoSwapConfig, Swap, SwapInfo } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAutoSwapsConfig() {
  return useSWR<AutoSwapConfig>("/api/autoswap", swrFetcher);
}

export function useSwapInfo(direction: "in" | "out") {
  return useSWR<SwapInfo>(`/api/swaps/${direction}/info`, swrFetcher);
}

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useSwap<T = Swap>(swapId: string | undefined, poll = false) {
  return useSWR<T>(
    swapId && `/api/swaps/${swapId}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
