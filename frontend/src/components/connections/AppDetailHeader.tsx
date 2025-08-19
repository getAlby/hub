import {
  CheckCircleIcon,
  EllipsisVerticalIcon,
  PlusCircleIcon,
  SquarePenIcon,
} from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { DisconnectApp } from "src/components/connections/DisconnectApp";
import { SuggestedApp } from "src/components/connections/SuggestedAppData";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { useAppsForAppStoreApp } from "src/hooks/useApps";
import { App } from "src/types";

export function AppDetailHeader({
  appStoreApp,
  contentRight,
}: {
  appStoreApp: SuggestedApp;
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
                className="w-14 h-14 rounded-lg mr-4"
              />
              <div className="flex flex-col">
                <div className="flex items-center gap-2">
                  {appStoreApp.title}
                  {!!connectedApps.length && (
                    <Badge
                      variant="positive"
                      className="flex items-center gap-1"
                    >
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
            <div className="hidden lg:block">
              <AppsActions
                appStoreApp={appStoreApp}
                connectedApps={connectedApps}
              />
            </div>
          )
        }
      />
      <div className="flex justify-end lg:hidden">
        <AppsActions appStoreApp={appStoreApp} connectedApps={connectedApps} />
      </div>
    </>
  );
}

function AppsActions({
  appStoreApp,
  connectedApps,
}: {
  appStoreApp: SuggestedApp;
  connectedApps: App[];
}) {
  return (
    <div className="flex items-center gap-2">
      {connectedApps.length === 0 ? (
        <Link to={`/apps/new?app=${appStoreApp.id}`}>
          <Button>
            <NostrWalletConnectIcon className="size-4" />
            Connect to {appStoreApp.title}
          </Button>
        </Link>
      ) : connectedApps.length === 1 ? (
        <SingleAppActions app={connectedApps[0]} appStoreApp={appStoreApp} />
      ) : (
        <MultipleAppActions appStoreApp={appStoreApp} />
      )}
    </div>
  );
}

function SingleAppActions({
  app,
  appStoreApp,
}: {
  app: App;
  appStoreApp: SuggestedApp;
}) {
  return (
    <>
      <MultipleAppActions appStoreApp={appStoreApp} />
      <DisconnectApp app={app} />
      <Link to={`/apps/${app.id}?edit=true`}>
        <Button variant="secondary">
          <SquarePenIcon className="size-4" /> Edit Connection
        </Button>
      </Link>
    </>
  );
}

function MultipleAppActions({ appStoreApp }: { appStoreApp: SuggestedApp }) {
  return (
    <>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger>
          <Button variant="outline" className="!px-2.5">
            <EllipsisVerticalIcon className="size-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          <DropdownMenuGroup>
            <DropdownMenuItem className="w-full">
              <Link
                to={`/apps/new?app=${appStoreApp.id}`}
                className="flex items-center gap-2"
              >
                <PlusCircleIcon className="w-4" /> Connect Again
              </Link>
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
