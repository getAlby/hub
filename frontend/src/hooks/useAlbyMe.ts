import useSWR from "swr";

import { AlbyMe } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAlbyMe() {
  return useSWR<AlbyMe>("/api/alby/me", swrFetcher, {
    dedupingInterval: 5 * 60 * 1000, // 5 minutes
  });
}
