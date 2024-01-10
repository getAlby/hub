import useSWR from "swr";

import { swrFetcher } from "@swr";
import { InfoResponse } from "@types";

export function useInfo() {
  return useSWR<InfoResponse>("/api/info", swrFetcher);
}
