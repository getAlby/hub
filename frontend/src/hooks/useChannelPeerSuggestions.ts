import useSWR from "swr";

import { RecommendedChannelPeer } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useChannelPeerSuggestions() {
  return useSWR<RecommendedChannelPeer[]>(
    "/api/channels/suggestions",
    swrFetcher
  );
}
