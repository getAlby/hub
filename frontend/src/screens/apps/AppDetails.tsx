import React from "react";
import { Link, useLocation, useParams } from "react-router-dom";

import {
  App,
  AppPermissions,
  UpdateAppRequest,
  WalletCapabilities,
} from "src/types";

import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear

import {
  CheckCircleIcon,
  ChevronDownIcon,
  EllipsisIcon,
  InfoIcon,
  PlusIcon,
  SquarePenIcon,
  SquareStackIcon,
  UnplugIcon,
} from "lucide-react";
import { toast } from "sonner";
import AppAvatar from "src/components/AppAvatar";
import AppHeader from "src/components/AppHeader";
import { AboutAppCard } from "src/components/connections/AboutAppCard";
import { AppLinksCard } from "src/components/connections/AppLinksCard";
import { AppTransactionList } from "src/components/connections/AppTransactionList";
import { AppUsage } from "src/components/connections/AppUsage";
import { ConnectionDetailsModal } from "src/components/connections/ConnectionDetailsModal";
import { DisconnectApp } from "src/components/connections/DisconnectApp";
import { getAppStoreApp } from "src/components/connections/SuggestedAppData";
import Loading from "src/components/Loading";
import Permissions from "src/components/Permissions";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { Input } from "src/components/ui/input";
import {
  ALBY_ACCOUNT_APP_NAME,
  SUBWALLET_APPSTORE_APP_ID,
} from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApp } from "src/hooks/useApp";
import { useAppsForAppStoreApp } from "src/hooks/useApps";
import { useCapabilities } from "src/hooks/useCapabilities";
import { cn } from "src/lib/utils";

function AppDetails() {
  const { id } = useParams() as { id: string };
  const { data: app, mutate: refetchApp, error } = useApp(parseInt(id));
  const { data: capabilities } = useCapabilities();

  if (error) {
    return <p className="text-red-500">{error.message}</p>;
  }

  if (!app || !capabilities) {
    return <Loading />;
  }

  return (
    <AppInternal
      app={app}
      refetchApp={refetchApp}
      capabilities={capabilities}
    />
  );
}

type AppInternalProps = {
  app: App;
  capabilities: WalletCapabilities;
  refetchApp: () => void;
};

function AppInternal({ app, refetchApp, capabilities }: AppInternalProps) {
  const location = useLocation();
  const [isEditingPermissions, setIsEditingPermissions] = React.useState(false);
  const [showConnectionDetails, setShowConnectionDetails] =
    React.useState(false);
  const [showDisconnectAppDialog, setShowDisconnectAppDialog] =
    React.useState(false);

  const { data: albyMe } = useAlbyMe();

  React.useEffect(() => {
    const queryParams = new URLSearchParams(location.search);
    const editMode = queryParams.has("edit");
    setIsEditingPermissions(editMode);
  }, [location.search]);

  const [name, setName] = React.useState(app.name);
  const [permissions, setPermissions] = React.useState<AppPermissions>({
    scopes: app.scopes,
    maxAmount: app.maxAmount,
    budgetRenewal: app.budgetRenewal,
    expiresAt: app.expiresAt ? new Date(app.expiresAt) : undefined,
    isolated: app.isolated,
  });
  const [savedPermissions, setSavedPermissions] =
    React.useState<AppPermissions>(permissions);

  const handleSave = async () => {
    try {
      const updateAppRequest: UpdateAppRequest = {
        name,
        scopes: Array.from(permissions.scopes),
        budgetRenewal: permissions.budgetRenewal,
        expiresAt: permissions.expiresAt?.toISOString(),
        maxAmount: permissions.maxAmount,
        isolated: permissions.isolated,
      };

      await request(`/api/apps/${app.appPubkey}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateAppRequest),
      });

      refetchApp();
      setIsEditingPermissions(false);
      setSavedPermissions(permissions);
      toast("Successfully updated connection");
    } catch (error) {
      handleRequestError("Failed to update connection", error);
    }
  };

  const handleConvertToSubwallet = async () => {
    try {
      const updateAppRequest: UpdateAppRequest = {
        name: app.name,
        scopes: app.scopes,
        budgetRenewal: app.budgetRenewal,
        expiresAt: app.expiresAt,
        maxAmount: app.maxAmount,
        isolated: app.isolated,
        metadata: {
          ...app.metadata,
          app_store_app_id: SUBWALLET_APPSTORE_APP_ID,
        },
      };

      await request(`/api/apps/${app.appPubkey}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateAppRequest),
      });

      refetchApp();
      toast("Successfully converted to sub-wallet", {
        description: "This isolated app is now a sub-wallet.",
      });
    } catch (error) {
      handleRequestError("Failed to convert to sub-wallet", error);
    }
  };

  const appName =
    app.name === ALBY_ACCOUNT_APP_NAME ? "Alby Account" : app.name;

  const appStoreApp = getAppStoreApp(app);
  const connectedApps = useAppsForAppStoreApp(appStoreApp);

  return (
    <>
      <div className="w-full">
        <div className="flex flex-col gap-2">
          <AppHeader
            title={
              <div className="flex flex-row gap-2 items-center">
                <AppAvatar app={app} className="w-10 h-10" />
                <h2
                  title={appName}
                  className="text-xl font-semibold overflow-hidden text-ellipsis whitespace-nowrap"
                >
                  {appName}
                </h2>
                <Badge variant="positive" className="flex items-center gap-1">
                  {(connectedApps?.length || 0) > 1 ? (
                    <DropdownMenu
                      modal={false}
                      key={JSON.stringify(app) /* force reload on app change */}
                    >
                      <DropdownMenuTrigger>
                        <div className="flex items-center gap-1">
                          {`${connectedApps?.length} Connections`}{" "}
                          <ChevronDownIcon className="size-3 -mr-1" />
                        </div>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-56">
                        <DropdownMenuGroup>
                          {connectedApps?.map((connectedApp) => (
                            <DropdownMenuItem key={connectedApp.id}>
                              <Link
                                to={`/apps/${connectedApp.id}`}
                                className={cn(
                                  "flex flex-1 items-center gap-2",
                                  connectedApp.id === app.id && "font-semibold"
                                )}
                              >
                                {connectedApp.name}
                              </Link>
                            </DropdownMenuItem>
                          ))}
                        </DropdownMenuGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  ) : (
                    <>
                      <CheckCircleIcon className="w-3 h-3" /> Connected
                    </>
                  )}
                </Badge>
              </div>
            }
            contentRight={
              <div className="flex gap-2 items-center">
                {!isEditingPermissions && (
                  <>
                    <DropdownMenu modal={false}>
                      <DropdownMenuTrigger>
                        <Button variant="outline" size="icon">
                          <EllipsisIcon />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {app.isolated &&
                          !app.metadata?.app_store_app_id &&
                          albyMe?.subscription.plan_code && (
                            <DropdownMenuGroup>
                              <DropdownMenuItem
                                onClick={handleConvertToSubwallet}
                              >
                                <SquareStackIcon /> Convert to Sub-wallet
                              </DropdownMenuItem>
                            </DropdownMenuGroup>
                          )}
                        <DropdownMenuGroup>
                          {appStoreApp && (
                            <DropdownMenuItem className="w-full">
                              <Link
                                to={`/apps/new?app=${appStoreApp.id}`}
                                className="flex flex-1 items-center gap-2"
                              >
                                <PlusIcon className="size-4" /> Add Another
                                Connection
                              </Link>
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuItem className="w-full">
                            <div
                              className="flex items-center gap-2"
                              onClick={() => setShowConnectionDetails(true)}
                            >
                              <InfoIcon className="size-4" /> More Connection
                              Details
                            </div>
                          </DropdownMenuItem>
                          <DropdownMenuItem className="w-full">
                            <div
                              className="flex items-center gap-2"
                              onClick={() => setShowDisconnectAppDialog(true)}
                            >
                              <UnplugIcon className="size-4" /> Disconnect{" "}
                              {appName}
                            </div>
                          </DropdownMenuItem>
                        </DropdownMenuGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                    <Button
                      variant="secondary"
                      onClick={() => setIsEditingPermissions(true)}
                    >
                      <SquarePenIcon className="size-4" /> Edit Connection
                    </Button>
                  </>
                )}
                {isEditingPermissions && (
                  <>
                    {isEditingPermissions && (
                      <div className="flex justify-center items-center gap-2">
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => {
                            setIsEditingPermissions(false);
                          }}
                        >
                          Cancel
                        </Button>

                        {(app.isolated && !permissions.isolated) ||
                        (!app.scopes.includes("pay_invoice") &&
                          permissions.scopes.includes("pay_invoice")) ? (
                          <AlertDialog>
                            <AlertDialogTrigger>
                              <Button type="button">Save</Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogTitle>
                                Confirm Update App
                              </AlertDialogTitle>
                              <AlertDialogDescription>
                                <div className="space-y-2">
                                  {app.isolated && !permissions.isolated ? (
                                    <p>
                                      Are you sure you wish to remove the{" "}
                                      <span className="font-bold">
                                        isolated
                                      </span>{" "}
                                      status from this connection?
                                    </p>
                                  ) : (
                                    <p>
                                      Are you sure you wish to give this
                                      connection{" "}
                                      <span className="font-bold">
                                        pay permissions
                                      </span>
                                      ?
                                    </p>
                                  )}
                                  <p className="text-amber-600 dark:text-amber-400 font-medium">
                                    ⚠️ Warning: This applies to all apps that
                                    have this connection secret. Only change
                                    this if you know it is safe to do so,
                                    otherwise you could potentially lose all
                                    funds
                                    {!!permissions.maxAmount &&
                                      " up to the specified budget"}
                                    {permissions.isolated &&
                                      " that are deposited into this isolated app"}
                                    .
                                  </p>
                                </div>
                              </AlertDialogDescription>
                              <AlertDialogFooter className="mt-5">
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <Button onClick={handleSave}>Save</Button>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        ) : (
                          <Button onClick={handleSave}>Save</Button>
                        )}
                      </div>
                    )}
                  </>
                )}
              </div>
            }
            description={""}
          />
          {!isEditingPermissions && (
            <>
              {appStoreApp && (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                  <AboutAppCard appStoreApp={appStoreApp} />
                  <AppLinksCard appStoreApp={appStoreApp} />
                </div>
              )}
              <AppUsage app={app} />
            </>
          )}
          {isEditingPermissions && app.name !== ALBY_ACCOUNT_APP_NAME && (
            <Card>
              <CardHeader>
                <CardTitle>App Name</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-row gap-2 items-center">
                  <Input
                    autoFocus
                    type="text"
                    name="name"
                    value={name}
                    id="name"
                    onChange={(e) => setName(e.target.value)}
                    required
                    autoComplete="off"
                  />
                </div>
              </CardContent>
            </Card>
          )}
          <Card>
            <CardHeader>
              <CardTitle>
                <div className="flex flex-row justify-between items-center">
                  Permissions
                </div>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Permissions
                capabilities={capabilities}
                permissions={
                  isEditingPermissions ? permissions : savedPermissions
                }
                setPermissions={setPermissions}
                readOnly={!isEditingPermissions}
                isNewConnection={false}
                budgetUsage={app.budgetUsage}
                showBudgetUsage={isEditingPermissions}
              />
            </CardContent>
          </Card>
          {!isEditingPermissions && (
            <>
              {showConnectionDetails && (
                <ConnectionDetailsModal
                  app={app}
                  onClose={() => setShowConnectionDetails(false)}
                />
              )}
              {showDisconnectAppDialog && (
                <DisconnectApp
                  app={app}
                  onClose={() => setShowDisconnectAppDialog(false)}
                />
              )}
              <AppTransactionList appId={app.id} />
            </>
          )}
        </div>
      </div>
    </>
  );
}

export default AppDetails;
