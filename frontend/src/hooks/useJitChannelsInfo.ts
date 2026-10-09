import useSWR from "swr";

import { JitChannelsInfoResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useJitChannelsInfo(enabled: boolean) {
  return useSWR<JitChannelsInfoResponse>(
    enabled ? "/api/jit-channels/info" : null,
    swrFetcher
  );
}
