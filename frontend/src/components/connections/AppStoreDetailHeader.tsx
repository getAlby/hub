import { CheckCircleIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Badge } from "src/components/ui/badge";
import { useAppsForAppStoreApp } from "src/hooks/useApps";

// TODO: remove once new connection wizard is added
export function AppStoreDetailHeader({
  appStoreApp,
  contentRight,
}: {
  appStoreApp: AppStoreApp;
  contentRight?: React.ReactNode | null;
}) {
  const connectedApps = useAppsForAppStoreApp(appStoreApp);
  if (!connectedApps) {
    return null;
  }

  return (
    <>
      <AppHeader
        title={
          <>
            <div className="flex flex-row items-center">
              <img
                src={appStoreApp.logo}
                className="w-14 h-14 rounded-lg mr-4 shrink-0"
              />
              <div className="flex gap-2 min-w-0">
                <h2
                  title={appStoreApp.title}
                  className="text-xl font-semibold overflow-hidden text-ellipsis whitespace-nowrap"
                >
                  {appStoreApp.title}
                </h2>
                {!!connectedApps.length && (
                  <Badge variant="positive" className="flex items-center gap-1">
                    <CheckCircleIcon className="w-3 h-3" />{" "}
                    {connectedApps.length > 1
                      ? `${connectedApps.length} Connections`
                      : "Connected"}
                  </Badge>
                )}
              </div>
            </div>
          </>
        }
        description=""
        contentRight={
          contentRight !== undefined ? (
            contentRight
          ) : (
            <Link to={`/apps/new?app=${appStoreApp.id}`}>
              <ResponsiveButton
                icon={NostrWalletConnectIcon}
                text={`Connect to ${appStoreApp.title}`}
              />
            </Link>
          )
        }
      />
    </>
  );
}
