import useSWR from "swr";

import React from "react";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import { LIST_APPS_LIMIT } from "src/constants";
import { ListAppsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useApps(
  limit = LIST_APPS_LIMIT,
  page = 1,
  filters?: {
    name?: string;
    appStoreAppId?: string;
    unused?: boolean;
    subWallets?: boolean;
  },
  orderBy?: "last_used_at" | "created_at",
  isEnabled = true
) {
  const offset = (page - 1) * limit;
  return useSWR<ListAppsResponse>(
    isEnabled
      ? `/api/apps?limit=${limit}&offset=${offset}&filters=${JSON.stringify(filters || {})}&order_by=${orderBy || ""}`
      : undefined,
    swrFetcher
  );
}

export function useAppsForAppStoreApp(appStoreApp: AppStoreApp | undefined) {
  const { data: connectedAppsByAppStoreId } = useApps(
    undefined,
    undefined,
    {
      appStoreAppId: appStoreApp?.id,
    },
    undefined,
    !!appStoreApp
  );
  const { data: connectedAppsByAppName } = useApps(
    undefined,
    undefined,
    {
      name: appStoreApp?.title,
    },
    undefined,
    !!appStoreApp
  );

  const connectedApps = React.useMemo(
    () =>
      connectedAppsByAppStoreId?.apps && connectedAppsByAppName?.apps
        ? [
            ...connectedAppsByAppStoreId.apps,
            ...connectedAppsByAppName.apps,
          ].filter((v, i, a) => a.findIndex((value) => value.id === v.id) === i)
        : undefined,
    [connectedAppsByAppName?.apps, connectedAppsByAppStoreId?.apps]
  );
  return connectedApps;
}
