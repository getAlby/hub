import { useApps } from "src/hooks/useApps";

export function useUnusedApps(limit?: number) {
  const { data: unusedAppsData } = useApps(
    limit,
    undefined,
    {
      unused: true,
    },
    undefined
  );
  return unusedAppsData?.apps;
}
