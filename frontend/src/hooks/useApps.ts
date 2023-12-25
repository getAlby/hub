import useSWR from "swr";
import { ListAppsResponse } from "../types";
import { swrFetcher } from "../swr";

export function useApps() {
  return useSWR<ListAppsResponse>("/api/apps", swrFetcher);
}
