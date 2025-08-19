import React from "react";
import { useLocation, useParams } from "react-router-dom";

import {
  App,
  AppPermissions,
  UpdateAppRequest,
  WalletCapabilities,
} from "src/types";

import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear

import {
  EllipsisIcon,
  LayoutGridIcon,
  PencilIcon,
  SquareStackIcon,
} from "lucide-react";
import AppAvatar from "src/components/AppAvatar";
import AppHeader from "src/components/AppHeader";
import { AppTransactionList } from "src/components/connections/AppTransactionList";
import { AppUsage } from "src/components/connections/AppUsage";
import { ConnectionSummary } from "src/components/connections/ConnectionSummary";
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
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { Input } from "src/components/ui/input";
import { useToast } from "src/components/ui/use-toast";
import {
  ALBY_ACCOUNT_APP_NAME,
  SUBWALLET_APPSTORE_APP_ID,
} from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApp } from "src/hooks/useApp";
import { useCapabilities } from "src/hooks/useCapabilities";

function ShowApp() {
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
  const { toast } = useToast();
  const location = useLocation();
  const [isEditingName, setIsEditingName] = React.useState(false);
  const [isEditingPermissions, setIsEditingPermissions] = React.useState(false);

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
      setIsEditingName(false);
      setIsEditingPermissions(false);
      setSavedPermissions(permissions);
      toast({
        title: "Successfully updated connection",
      });
    } catch (error) {
      handleRequestError(toast, "Failed to update connection", error);
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
      toast({
        title: "Successfully converted to sub-wallet",
        description: "This isolated app is now a sub-wallet.",
      });
    } catch (error) {
      handleRequestError(toast, "Failed to convert to sub-wallet", error);
    }
  };

  const appName =
    app.name === ALBY_ACCOUNT_APP_NAME ? "Alby Account" : app.name;

  const appStoreApp = getAppStoreApp(app);

  return (
    <>
      <div className="w-full">
        <div className="flex flex-col gap-5">
          <AppHeader
            title={
              <div className="flex flex-row items-center">
                <AppAvatar app={app} className="w-10 h-10 mr-2" />
                {isEditingName ? (
                  <div className="flex flex-row gap-2 items-center">
                    <Input
                      autoFocus
                      type="text"
                      name="name"
                      value={name}
                      id="name"
                      onChange={(e) => setName(e.target.value)}
                      required
                      className="text-xl font-semibold w-max max-w-40 md:max-w-fit"
                      autoComplete="off"
                    />
                    <Button type="button" onClick={handleSave}>
                      Save
                    </Button>
                  </div>
                ) : (
                  <div
                    className="flex flex-row gap-2 items-center cursor-pointer"
                    onClick={() => setIsEditingName(true)}
                  >
                    <h2
                      title={appName}
                      className="text-xl font-semibold overflow-hidden text-ellipsis whitespace-nowrap"
                    >
                      {appName}
                    </h2>
                    {app.name !== ALBY_ACCOUNT_APP_NAME && (
                      <PencilIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                    )}
                  </div>
                )}
              </div>
            }
            contentRight={
              <div className="flex gap-2 items-center">
                {app.isolated &&
                  !app.metadata?.app_store_app_id &&
                  albyMe?.subscription.plan_code && (
                    <DropdownMenu modal={false}>
                      <DropdownMenuTrigger>
                        <Button variant="outline" size="icon">
                          <EllipsisIcon />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-56">
                        <DropdownMenuGroup>
                          <DropdownMenuItem>
                            <div
                              className="w-full cursor-pointer flex items-center gap-2"
                              onClick={handleConvertToSubwallet}
                            >
                              <SquareStackIcon /> Convert to Sub-wallet
                            </div>
                          </DropdownMenuItem>
                        </DropdownMenuGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                <DisconnectApp app={app} />
                {appStoreApp && (
                  <LinkButton
                    variant="secondary"
                    to={
                      appStoreApp.internal
                        ? `/internal-apps/${appStoreApp.id}`
                        : `/appstore/${appStoreApp.id}`
                    }
                  >
                    <LayoutGridIcon /> View in App Store
                  </LinkButton>
                )}
              </div>
            }
            description={""}
          />
          <h2 className="font-semibold text-2xl">Manage Connection</h2>
          <Card className="gap-0">
            <CardHeader>
              <CardTitle>
                <div className="flex flex-row justify-between items-center">
                  Permissions
                  <div className="flex flex-row gap-2">
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
                            <AlertDialogTrigger asChild>
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

                    {!isEditingPermissions && (
                      <>
                        <Button
                          variant="outline"
                          onClick={() =>
                            setIsEditingPermissions(!isEditingPermissions)
                          }
                        >
                          Edit
                        </Button>
                      </>
                    )}
                  </div>
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
              />
            </CardContent>
          </Card>
          <ConnectionSummary app={app} />
          <AppUsage app={app} />
          <AppTransactionList appId={app.id} />
        </div>
      </div>
    </>
  );
}

export default ShowApp;
