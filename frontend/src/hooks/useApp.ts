import useSWR from "swr";

import { swrFetcher } from "@swr";
import { App, ErrorResponse } from "@types";

export function useApp(pubkey: string | undefined) {
  return useSWR<App | ErrorResponse>(
    pubkey && `/api/apps/${pubkey}`,
    swrFetcher
  );
}
