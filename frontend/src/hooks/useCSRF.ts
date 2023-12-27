import useSWR from "swr";
import { swrFetcher } from "../swr";

export function useCSRF() {
  return useSWR<string>("/api/csrf", swrFetcher);
}
