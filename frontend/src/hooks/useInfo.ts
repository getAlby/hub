import useSWR from "swr";
import { UserInfo } from "../types";
import { swrFetcher } from "../swr";

export function useInfo() {
  return useSWR<UserInfo>("/api/info", swrFetcher);
}
