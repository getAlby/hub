import useSWR from "swr";

import { useInfo } from "src/hooks/useInfo";
import { AlbyMe } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAlbyMe() {
  const { data: info } = useInfo();

  return useSWR<AlbyMe>(
    info?.albyAccountConnected ? "/api/alby/me" : undefined,
    swrFetcher,
    {
      dedupingInterval: 5 * 60 * 1000, // 5 minutes
    }
  );
}
