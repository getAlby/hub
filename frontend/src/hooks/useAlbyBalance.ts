import useSWR from "swr";

import { useInfo } from "src/hooks/useInfo";
import { AlbyBalance } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAlbyBalance() {
  const { data: info } = useInfo();
  return useSWR<AlbyBalance>(
    info?.albyAccountConnected ? "/api/alby/balance" : undefined,
    swrFetcher,
    {
      dedupingInterval: 5 * 60 * 1000, // 5 minutes
    }
  );
}
