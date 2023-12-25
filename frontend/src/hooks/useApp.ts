import useSWR from "swr";
import { ShowAppResponse } from "../types";
import { swrFetcher } from "../swr";

export function useApp(pubkey: string | undefined) {
  return useSWR<ShowAppResponse>(pubkey && `/api/apps/${pubkey}`, swrFetcher);
}
