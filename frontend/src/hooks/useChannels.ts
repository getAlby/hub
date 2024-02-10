import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { Channel } from "src/types";

export function useChannels() {
  return useSWR<Channel[]>("/api/channels", swrFetcher);
}
