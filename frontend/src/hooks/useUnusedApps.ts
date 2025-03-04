import dayjs from "dayjs";
import React from "react";
import { useApps } from "src/hooks/useApps";

const oldDate = dayjs().subtract(2, "months");
export function useUnusedApps() {
  const { data: apps } = useApps();

  return React.useMemo(
    () =>
      apps?.filter(
        (app) => !app.lastEventAt || dayjs(app.lastEventAt).isBefore(oldDate)
      ),
    [apps]
  );
}
