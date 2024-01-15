import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { App } from "src/types";

export function useApps() {
  return useSWR<App[]>("/api/apps", swrFetcher);
}
