import useSWR from "swr";
import { App, ErrorResponse } from "../types";
import { swrFetcher } from "../swr";

export function useApp(pubkey: string | undefined) {
  return useSWR<App | ErrorResponse>(
    pubkey && `/api/apps/${pubkey}`,
    swrFetcher
  );
}
