import useSWR from "swr";

import { WalletCapabilities } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useCapabilities() {
  return useSWR<WalletCapabilities>("/api/wallet/capabilities", swrFetcher);
}
