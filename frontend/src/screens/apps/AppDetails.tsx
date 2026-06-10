import React from "react";
import { Link, useLocation, useParams } from "react-router";

import {
  App,
  AppPermissions,
  BudgetRenewalType,
  Scope,
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
import { AppLinks } from "src/components/connections/AppLinksCard";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { AppTransactionList } from "src/components/connections/AppTransactionList";
import { AppUsage } from "src/components/connections/AppUsage";
import { ConnectionDetailsModal } from "src/components/connections/ConnectionDetailsModal";
import { DisconnectApp } from "src/components/connections/DisconnectApp";
import { getAppStoreApp } from "src/components/connections/SuggestedAppData";
import Loading from "src/components/Loading";
import Permissions from "src/components/Permissions";
import ResponsiveButton from "src/components/ResponsiveButton";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
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
  DropdownMenuSeparator,
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
import { cn, getAppDisplayName } from "src/lib/utils";

function formatList(items: string[]): string {
  if (items.length === 1) {
    return items[0];
  }
  if (items.length === 2) {
    return `${items[0]} and ${items[1]}`;
  }
  return `${items.slice(0, -1).join(", ")}, and ${items[items.length - 1]}`;
}

const budgetPeriodLabels: Record<BudgetRenewalType, string> = {
  daily: "day",
  weekly: "week",
  monthly: "month",
  yearly: "year",
  never: "",
  "": "",
};

function getSecondaryAbilities(scopes: Scope[]): string[] {
  const abilities: string[] = [];
  if (scopes.includes("make_invoice")) {
    abilities.push("receive payments");
  }
  if (
    scopes.includes("get_balance") ||
    scopes.includes("get_info") ||
    scopes.includes("list_transactions") ||
    scopes.includes("lookup_invoice")
  ) {
    abilities.push("read your balance and activity");
  }
  if (scopes.includes("sign_message")) {
    abilities.push("sign messages");
  }
  return abilities;
}

function PermissionsSummary({
  scopes,
  maxAmountSat,
  budgetRenewal,
}: {
  scopes: Scope[];
  maxAmountSat: number;
  budgetRenewal: BudgetRenewalType;
}) {
  if (scopes.includes("superuser")) {
    return (
      <p>
        This app has full access and can create other connections to your
        wallet.
      </p>
    );
  }

  const canSpend = scopes.includes("pay_invoice");
  const secondary = getSecondaryAbilities(scopes);

  if (!canSpend && !secondary.length) {
    return <p>This app has no access to your wallet.</p>;
  }

  if (!canSpend) {
    return <p>This app can {formatList(secondary)}.</p>;
  }

  return (
    <p>
      This app can send payments
      {maxAmountSat > 0 ? (
        <>
          {" "}
          up to{" "}
          <span className="font-medium text-foreground">
            <FormattedBitcoinAmount amountMsat={maxAmountSat * 1000} />
          </span>
          {budgetRenewal !== "never" && (
            <> per {budgetPeriodLabels[budgetRenewal]}</>
          )}
        </>
      ) : (
        <> with no spending limit set</>
      )}
      .{!!secondary.length && <> It can also {formatList(secondary)}.</>}
    </p>
  );
}

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
  const [showConfirmDialog, setShowConfirmDialog] = React.useState(false);
  const [showConnectionDetails, setShowConnectionDetails] =
    React.useState(false);
  const [showDisconnectAppDialog, setShowDisconnectAppDialog] =
    React.useState(false);
  const [showPermissionDetails, setShowPermissionDetails] =
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
    maxAmountSat: app.maxAmountSat,
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
        updateExpiresAt: true,
        maxAmountSat: permissions.maxAmountSat,
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
      // Only send the metadata since that's the only thing changing
      const updateAppRequest: UpdateAppRequest = {
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

  const appName = getAppDisplayName(app.name);

  const appStoreApp = getAppStoreApp(app);
  const connectedApps = useAppsForAppStoreApp(appStoreApp);
  const needsConfirmation =
    (app.isolated && !permissions.isolated) ||
    (!app.scopes.includes("pay_invoice") &&
      permissions.scopes.includes("pay_invoice"));

  const permissionsSection = (
    <div>
      <div className="flex flex-row justify-between items-center mb-2">
        <p className="font-semibold">Permissions</p>
        <button
          type="button"
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          onClick={() => setShowPermissionDetails((current) => !current)}
        >
          {showPermissionDetails ? "Hide details" : "Show details"}
          <ChevronDownIcon
            className={cn(
              "size-4 transition-transform",
              showPermissionDetails && "rotate-180"
            )}
          />
        </button>
      </div>
      {showPermissionDetails ? (
        <Permissions
          capabilities={capabilities}
          permissions={savedPermissions}
          setPermissions={setPermissions}
          readOnly
        />
      ) : (
        <div className="text-sm text-muted-foreground space-y-1">
          <PermissionsSummary
            scopes={app.scopes}
            maxAmountSat={app.maxAmountSat}
            budgetRenewal={app.budgetRenewal}
          />
          {app.expiresAt && (
            <p>Expires {new Date(app.expiresAt).toLocaleDateString()}</p>
          )}
        </div>
      )}
    </div>
  );

  return (
    <>
      <div className="w-full">
        <div className="flex flex-col gap-2">
          <AppHeader
            pageTitle={appName}
            title={
              <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
                <div className="flex flex-row gap-2 items-center min-w-0">
                  <AppAvatar app={app} className="w-10 h-10 shrink-0" />
                  <h2
                    title={appName}
                    className="text-xl font-semibold overflow-hidden text-ellipsis whitespace-nowrap"
                  >
                    {appName}
                  </h2>
                </div>
                <Badge
                  variant="positive"
                  className="flex items-center gap-1 self-start sm:self-center"
                >
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
                      <Button variant="outline" size="icon" asChild>
                        <DropdownMenuTrigger>
                          <EllipsisIcon />
                        </DropdownMenuTrigger>
                      </Button>
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
                            <DropdownMenuItem asChild>
                              <Link
                                to={`/apps/new?app=${appStoreApp.id}`}
                                className="flex flex-1 items-center gap-2"
                              >
                                <PlusIcon className="size-4" /> Add Another
                                Connection
                              </Link>
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuItem asChild>
                            <div
                              className="flex items-center gap-2"
                              onClick={() => setShowConnectionDetails(true)}
                            >
                              <InfoIcon className="size-4" /> Connection Details
                            </div>
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem variant="destructive" asChild>
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
                    <ResponsiveButton
                      variant="secondary"
                      onClick={() => setIsEditingPermissions(true)}
                      icon={SquarePenIcon}
                      text="Edit Connection"
                    />
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

                        {needsConfirmation ? (
                          <AlertDialog
                            open={showConfirmDialog}
                            onOpenChange={setShowConfirmDialog}
                          >
                            <Button type="submit" form="app-details">
                              Save
                            </Button>
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
                                    {!!permissions.maxAmountSat &&
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
                          <Button type="submit" form="app-details">
                            Save
                          </Button>
                        )}
                      </div>
                    )}
                  </>
                )}
              </div>
            }
            description={""}
          />
          {isEditingPermissions ? (
            <form
              id="app-details"
              className="flex flex-col gap-2"
              onSubmit={(e) => {
                e.preventDefault();
                if (needsConfirmation) {
                  setShowConfirmDialog(true);
                  return;
                }
                handleSave();
              }}
            >
              {app.name !== ALBY_ACCOUNT_APP_NAME && (
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
                    permissions={permissions}
                    setPermissions={setPermissions}
                  />
                </CardContent>
              </Card>
            </form>
          ) : (
            <div
              className={cn(
                "grid grid-cols-1 gap-3 items-start",
                appStoreApp && "lg:grid-cols-3"
              )}
            >
              <div
                className={cn(
                  "flex flex-col gap-3 min-w-0",
                  appStoreApp && "lg:col-span-2"
                )}
              >
                <AppUsage app={app} />
                {!appStoreApp && (
                  <Card>
                    <CardContent className="py-6">
                      {permissionsSection}
                    </CardContent>
                  </Card>
                )}
                <AppTransactionList appId={app.id} />
              </div>
              {appStoreApp && (
                <Card className="gap-4">
                  <CardHeader>
                    <CardTitle>About the App</CardTitle>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-4">
                    <p className="text-muted-foreground">
                      {appStoreApp.extendedDescription}
                    </p>
                    <AppLinks appStoreApp={appStoreApp} />
                    <div className="border-t pt-4">{permissionsSection}</div>
                  </CardContent>
                </Card>
              )}
            </div>
          )}
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
            </>
          )}
        </div>
      </div>
    </>
  );
}

export default AppDetails;
