import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { AlbyBalance } from "src/types";

export function useAlbyBalance() {
  return useSWR<AlbyBalance>("/api/alby/balance", swrFetcher);
}
