import useSWR from "swr";

import { swrFetcher } from "src/swr";

export function useCSRF() {
  return useSWR<string>("/api/csrf", swrFetcher);
}
