import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { OnchainBalanceResponse } from "src/types";

export function useOnchainBalance() {
  return useSWR<OnchainBalanceResponse>("/api/wallet/balance", swrFetcher);
}
