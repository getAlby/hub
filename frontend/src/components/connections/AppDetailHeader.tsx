import {
  CheckCircleIcon,
  EllipsisVerticalIcon,
  PlusCircleIcon,
  SquarePenIcon,
  UnplugIcon,
} from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
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
        ) : connectedApps.length === 0 ? (
          <Link to={`/apps/new?app=${appStoreApp.id}`}>
            <Button>
              <NostrWalletConnectIcon className="size-4" />
              Connect to {appStoreApp.title}
            </Button>
          </Link>
        ) : connectedApps.length === 1 ? (
          <SingleAppActions app={connectedApps[0]} />
        ) : (
          <MultipleAppActions />
        )
      }
    />
  );
}

function SingleAppActions({ app }: { app: App }) {
  return (
    <>
      <MultipleAppActions />
      <Button variant="outline" onClick={() => alert("TODO")}>
        <UnplugIcon className="size-4" /> Disconnect
      </Button>
      <Button variant="secondary" onClick={() => alert("TODO")}>
        <SquarePenIcon className="size-4" /> Edit Connection
      </Button>
    </>
  );
}

function MultipleAppActions() {
  return (
    <>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger>
          <EllipsisVerticalIcon className="size-4" />
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          <DropdownMenuGroup>
            <DropdownMenuItem className="w-full">
              <div
                className="w-full cursor-pointer flex items-center gap-2"
                onClick={() => alert("TODO")}
              >
                <PlusCircleIcon className="w-4" /> Connect Again
              </div>
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
