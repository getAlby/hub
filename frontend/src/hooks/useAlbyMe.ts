import useSWR from "swr";

import { useInfo } from "src/hooks/useInfo";
import { AlbyMe } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAlbyMe() {
  const { data: info, isLoading: isInfoLoading } = useInfo();

  const albyMe = useSWR<AlbyMe>(
    info?.albyAccountConnected ? "/api/alby/me" : undefined,
    swrFetcher,
    {
      dedupingInterval: 5 * 60 * 1000, // 5 minutes
    }
  );
  const isAlbyLoading = info?.albyAccountConnected ? albyMe.isLoading : false;

  return {
    ...albyMe,
    isLoading: isInfoLoading || isAlbyLoading,
  };
}
