import useSWR, { SWRConfiguration } from "swr";

import React from "react";
import { backendTypeConfigs } from "src/lib/backendType";
import { InfoResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useInfo(poll = false) {
  const info = useSWR<InfoResponse>(
    "/api/info",
    swrFetcher,
    poll ? pollConfiguration : undefined
  );

  return React.useMemo(
    () => ({
      ...info,
      hasChannelManagement:
        info.data?.backendType &&
        backendTypeConfigs[info.data.backendType].hasChannelManagement,
      hasMnemonic:
        info.data?.backendType &&
        backendTypeConfigs[info.data.backendType].hasMnemonic,
      hasNodeBackup:
        info.data?.backendType &&
        backendTypeConfigs[info.data.backendType].hasNodeBackup,
    }),
    [info]
  );
}
