import { CheckCircleIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { SuggestedApp } from "src/components/connections/SuggestedAppData";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import { useAppsForAppStoreApp } from "src/hooks/useApps";

export function AppDetailHeader({
  appStoreApp,
  contentRight,
}: {
  appStoreApp: SuggestedApp;
  contentRight?: React.ReactNode | null;
}) {
  const connectedApps = useAppsForAppStoreApp(appStoreApp);

  return (
    <AppHeader
      title={
        <>
          <div className="flex flex-row items-center">
            <img src={appStoreApp.logo} className="w-14 h-14 rounded-lg mr-4" />
            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                {appStoreApp.title}
                {!!connectedApps.length && (
                  <Badge variant="positive" className="flex items-center gap-1">
                    <CheckCircleIcon className="w-3 h-3" />{" "}
                    {connectedApps.length > 1
                      ? `${connectedApps.length} Connections`
                      : "Connected"}
                  </Badge>
                )}
              </div>
              <div className="text-sm font-normal text-muted-foreground">
                {appStoreApp.description}
              </div>
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
            <Button>
              <NostrWalletConnectIcon className="w-4 h-4 mr-2" />
              Connect to {appStoreApp.title}
            </Button>
          </Link>
        )
      }
    />
  );
}
