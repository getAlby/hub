import AppCard from "src/components/connections/AppCard";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import { useAppsForAppStoreApp } from "src/hooks/useApps";

export function AppDetailConnectedApps({
  appStoreApp,
  showTitle,
}: {
  appStoreApp: AppStoreApp;
  showTitle?: boolean;
}) {
  const connectedApps = useAppsForAppStoreApp(appStoreApp);

  if (!connectedApps?.length) {
    return null;
  }

  return (
    <>
      {showTitle && (
        <h2 className="font-medium text-lg mt-6">Your Connections</h2>
      )}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {connectedApps.map((app) => (
          <AppCard key={app.id} app={app} />
        ))}
      </div>
    </>
  );
}
