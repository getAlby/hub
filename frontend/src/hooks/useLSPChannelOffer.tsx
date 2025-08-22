import useSWR from "swr";

import { LSPChannelOffer } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useLSPChannelOffer() {
  return useSWR<LSPChannelOffer>("/api/channel-offer", swrFetcher);
}
