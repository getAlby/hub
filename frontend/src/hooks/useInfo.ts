import useSWR from "swr";

import { swrFetcher } from "src/swr";
import { InfoResponse } from "src/types";

export function useInfo() {
  return useSWR<InfoResponse>("/api/info", swrFetcher);
}
