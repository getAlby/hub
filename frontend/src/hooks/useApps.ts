import useSWR from "swr";

import { swrFetcher } from "@swr";
import { App } from "@types";

export function useApps() {
  return useSWR<App[]>("/api/apps", swrFetcher);
}
