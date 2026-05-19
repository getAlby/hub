import useSWR from "swr";

import { LSPChannelOffer } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useLSPChannelOffer(enabled = true) {
  return useSWR<LSPChannelOffer>(
    enabled ? "/api/channel-offer" : undefined,
    swrFetcher
  );
}
