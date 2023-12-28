import useSWR from "swr";
import { App } from "../types";
import { swrFetcher } from "../swr";

export function useApps() {
  return useSWR<App[]>("/api/apps", swrFetcher);
}
