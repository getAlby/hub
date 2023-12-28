import useSWR from "swr";
import { InfoResponse } from "../types";
import { swrFetcher } from "../swr";

export function useInfo() {
  return useSWR<InfoResponse>("/api/info", swrFetcher);
}
