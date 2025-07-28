import useSWR, { SWRConfiguration } from "swr";

import { App } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useAppByPubkey(pubkey: string | undefined, poll = false) {
  return useSWR<App>(
    pubkey && `/api/apps/${pubkey}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}

export function useApp(id: number | undefined, poll = false) {
  return useSWR<App>(
    !!id && `/api/v2/apps/${id}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}
