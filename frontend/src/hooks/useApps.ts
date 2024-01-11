import useSWR from "swr";

import { swrFetcher } from "src/swr";
import { App } from "src/types";

export function useApps() {
  return useSWR<App[]>("/api/apps", swrFetcher);
}
