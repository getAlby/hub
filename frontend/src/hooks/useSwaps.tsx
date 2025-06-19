import useSWR, { SWRConfiguration } from "swr";

import { AutoSwapsConfig, Swap, SwapFees } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAutoSwapsConfig() {
  return useSWR<AutoSwapsConfig>("/api/wallet/autoswap", swrFetcher);
}

export function useSwapFees(direction: "in" | "out") {
  return useSWR<SwapFees>(`/api/wallet/swap/${direction}/fees`, swrFetcher);
}

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useSwap(swapId: string, poll = false) {
  return useSWR<Swap>(
    `/api/wallet/swap/info/${swapId}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
