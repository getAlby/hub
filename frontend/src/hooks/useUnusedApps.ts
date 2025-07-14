import dayjs from "dayjs";
import React from "react";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useApps } from "src/hooks/useApps";

const OLD_DATE = dayjs().subtract(2, "months");

export function useUnusedApps() {
  const { data: apps } = useApps();

  return React.useMemo(
    () =>
      apps?.filter(
        (app) =>
          (!app.lastUsedAt || dayjs(app.lastUsedAt).isBefore(OLD_DATE)) &&
          app.metadata?.app_store_app_id != SUBWALLET_APPSTORE_APP_ID
      ),
    [apps]
  );
}
