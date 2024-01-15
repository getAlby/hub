import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { App } from "src/types";

export function useApp(pubkey: string | undefined) {
  return useSWR<App>(pubkey && `/api/apps/${pubkey}`, swrFetcher);
}
