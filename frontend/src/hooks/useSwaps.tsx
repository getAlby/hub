import useSWR from "swr";

import { AutoSwapsConfig, SwapFees } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAutoSwapsConfig() {
  return useSWR<AutoSwapsConfig>("/api/wallet/autoswap", swrFetcher);
}

export function useSwapFees(direction: "in" | "out") {
  return useSWR<SwapFees>(`/api/wallet/swap/${direction}/fees`, swrFetcher);
}
