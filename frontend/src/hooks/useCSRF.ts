import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";

export function useCSRF() {
  return useSWR<string>("/api/csrf", swrFetcher);
}
