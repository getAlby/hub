import { useApps } from "src/hooks/useApps";

export function useRiskyApps(limit?: number) {
  const { data: riskyAppsData } = useApps(
    limit,
    undefined,
    {
      risky: true,
      subWallets: false,
    },
    undefined
  );
  return riskyAppsData;
}
