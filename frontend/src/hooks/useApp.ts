import useSWR from "swr";

import { swrFetcher } from "src/swr";
import { App, ErrorResponse } from "src/types";

export function useApp(pubkey: string | undefined) {
  return useSWR<App | ErrorResponse>(
    pubkey && `/api/apps/${pubkey}`,
    swrFetcher
  );
}
