import useSWR from "swr";

import React from "react";
import { SuggestedApp } from "src/components/connections/SuggestedAppData";
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
  },
  orderBy?: "last_used_at" | "created_at"
) {
  const offset = (page - 1) * limit;
  return useSWR<ListAppsResponse>(
    `/api/apps?limit=${limit}&offset=${offset}&filters=${JSON.stringify(filters || {})}&order_by=${orderBy || ""}`,
    swrFetcher
  );
}

export function useAppsForAppStoreApp(appStoreApp: SuggestedApp) {
  const { data: connectedAppsByAppStoreId } = useApps(undefined, undefined, {
    appStoreAppId: appStoreApp.id,
  });
  const { data: connectedAppsByAppName } = useApps(undefined, undefined, {
    name: appStoreApp.title,
  });

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
