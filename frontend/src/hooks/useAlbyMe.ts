import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { AlbyMe } from "src/types";

export function useAlbyMe() {
  return useSWR<AlbyMe>("/api/alby/me", swrFetcher);
}
