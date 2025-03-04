import dayjs from "dayjs";
import { useApps } from "src/hooks/useApps";

export function useUnusedApps() {
  const { data: apps } = useApps();

  const oldDate = dayjs().subtract(2, "months");
  return apps?.filter(
    (app) => !app.lastEventAt || dayjs(app.lastEventAt).isBefore(oldDate)
  );
}
