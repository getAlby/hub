import useSWR from "swr";
import { App } from "../types";
import { swrFetcher } from "../swr";

export function useApp(pubkey: string | undefined) {
  return useSWR<App>(pubkey && `/api/apps/${pubkey}`, swrFetcher);
}
