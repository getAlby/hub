import AppCard from "src/components/connections/AppCard";
import { SuggestedApp } from "src/components/connections/SuggestedAppData";
import { useAppsForAppStoreApp } from "src/hooks/useApps";

export function AppDetailConnectedApps({
  appStoreApp,
}: {
  appStoreApp: SuggestedApp;
}) {
  const connectedApps = useAppsForAppStoreApp(appStoreApp);

  if (!connectedApps?.length) {
    return null;
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {connectedApps.map((app) => (
        <AppCard key={app.id} app={app} />
      ))}
    </div>
  );
}
