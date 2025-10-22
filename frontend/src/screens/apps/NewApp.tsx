import { useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";

import React from "react";
import { toast } from "sonner";
import Loading from "src/components/Loading";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import PasswordInput from "src/components/password/PasswordInput";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { useCapabilities } from "src/hooks/useCapabilities";
import { createApp } from "src/requests/createApp";
import {
  AppPermissions,
  BudgetRenewalType,
  CreateAppRequest,
  CreateAppResponse,
  Nip47NotificationType,
  Nip47RequestMethod,
  Scope,
  WalletCapabilities,
  validBudgetRenewals,
} from "src/types";

import AppHeader from "src/components/AppHeader";
import { IsolatedAppTopupDialog } from "src/components/IsolatedAppTopupDialog";
import { InstallApp } from "src/components/connections/InstallApp";
import { defineStepper } from "src/components/stepper";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { useApp } from "src/hooks/useApp";
import { ConnectAppCard } from "src/screens/apps/ConnectAppCard";
import { handleRequestError } from "src/utils/handleRequestError";
import Permissions from "../../components/Permissions";
import { AppStoreApp } from "../../components/connections/SuggestedAppData";

const NewApp = () => {
  const { data: capabilities } = useCapabilities();
  if (!capabilities) {
    return <Loading />;
  }

  return <NewAppInternal capabilities={capabilities} />;
};

type NewAppInternalProps = {
  capabilities: WalletCapabilities;
};

const NewAppInternal = ({ capabilities }: NewAppInternalProps) => {
  const location = useLocation();

  const [unsupportedError, setUnsupportedError] = useState<string>();
  const [isLoading, setLoading] = React.useState(false);

  const [createAppResponse, setCreateAppResponse] =
    React.useState<CreateAppResponse>();

  const queryParams = new URLSearchParams(location.search);

  const appId = queryParams.get("app") ?? "";
  const appStoreApp = appStoreApps.find((app) => app.id === appId);
  const isInstallable =
    appStoreApp?.appleLink ||
    appStoreApp?.playLink ||
    appStoreApp?.zapStoreLink ||
    appStoreApp?.chromeLink ||
    appStoreApp?.firefoxLink;

  const pubkey = queryParams.get("pubkey") ?? "";
  const returnTo = queryParams.get("return_to") ?? "";

  const nameParam = (queryParams.get("name") || queryParams.get("c")) ?? "";
  const [appName, setAppName] = useState(
    appStoreApp ? appStoreApp.title : nameParam
  );

  const budgetRenewalParam = queryParams.get(
    "budget_renewal"
  ) as BudgetRenewalType;
  const budgetMaxAmountMsatParam = queryParams.get("max_amount") ?? "";
  const isolatedParam = queryParams.get("isolated") ?? "";
  const expiresAtParam = queryParams.get("expires_at") ?? "";

  const reqMethodsParam = queryParams.get("request_methods") ?? "";
  const notificationTypesParam = queryParams.get("notification_types") ?? "";

  const initialScopes: Scope[] = React.useMemo(() => {
    const methods = reqMethodsParam
      ? reqMethodsParam.split(" ")
      : capabilities.methods;

    const requestMethodsSet = new Set<Nip47RequestMethod>(
      methods as Nip47RequestMethod[]
    );
    const unsupportedMethods = Array.from(requestMethodsSet).filter(
      (method) => capabilities.methods.indexOf(method) < 0
    );
    if (unsupportedMethods.length) {
      setUnsupportedError(
        "This app requests methods not supported by your wallet: " +
          unsupportedMethods
      );
    }

    const notificationTypes = notificationTypesParam
      ? notificationTypesParam.split(" ")
      : reqMethodsParam
        ? [] // do not set notifications if only request methods provided
        : capabilities.notificationTypes;

    const notificationTypesSet = new Set<Nip47NotificationType>(
      notificationTypes as Nip47NotificationType[]
    );
    const unsupportedNotificationTypes = Array.from(
      notificationTypesSet
    ).filter(
      (notificationType) =>
        capabilities.notificationTypes.indexOf(notificationType) < 0
    );
    if (unsupportedNotificationTypes.length) {
      setUnsupportedError(
        "This app requests notification types not supported by your wallet: " +
          unsupportedNotificationTypes
      );
    }

    const scopes: Scope[] = [];
    if (
      requestMethodsSet.has("pay_invoice") ||
      requestMethodsSet.has("pay_keysend") ||
      requestMethodsSet.has("multi_pay_invoice") ||
      requestMethodsSet.has("multi_pay_keysend")
    ) {
      scopes.push("pay_invoice");
    }

    if (requestMethodsSet.has("get_info")) {
      scopes.push("get_info");
    }
    if (requestMethodsSet.has("get_balance")) {
      scopes.push("get_balance");
    }
    if (
      requestMethodsSet.has("make_invoice") ||
      requestMethodsSet.has("make_hold_invoice") ||
      requestMethodsSet.has("settle_hold_invoice") ||
      requestMethodsSet.has("cancel_hold_invoice")
    ) {
      scopes.push("make_invoice");
    }
    if (requestMethodsSet.has("lookup_invoice")) {
      scopes.push("lookup_invoice");
    }
    if (requestMethodsSet.has("list_transactions")) {
      scopes.push("list_transactions");
    }
    if (requestMethodsSet.has("sign_message") && isolatedParam !== "true") {
      scopes.push("sign_message");
    }
    if (notificationTypes.length) {
      scopes.push("notifications");
    }

    return scopes;
  }, [
    capabilities.methods,
    capabilities.notificationTypes,
    isolatedParam,
    notificationTypesParam,
    reqMethodsParam,
  ]);

  const parseExpiresParam = (expiresParam: string): Date | undefined => {
    const expiresParamTimestamp = parseInt(expiresParam);
    if (!isNaN(expiresParamTimestamp)) {
      const expiry = new Date(expiresParamTimestamp * 1000);
      expiry.setHours(23, 59, 59);
      return expiry;
    }
    return undefined;
  };

  const [superuser, setSuperuser] = useState(appStoreApp?.superuser || false);
  const [
    showSuperuserConfirmPasswordDialog,
    setShowSuperuserConfirmPasswordDialog,
  ] = useState(false);
  const [unlockPassword, setUnlockPassword] = useState("");

  const [permissions, setPermissions] = useState<AppPermissions>({
    scopes: initialScopes,
    maxAmount: budgetMaxAmountMsatParam
      ? Math.floor(parseInt(budgetMaxAmountMsatParam) / 1000)
      : 0,
    budgetRenewal: validBudgetRenewals.includes(budgetRenewalParam)
      ? budgetRenewalParam
      : budgetMaxAmountMsatParam
        ? "never"
        : "monthly",
    expiresAt: parseExpiresParam(expiresAtParam),
    isolated: isolatedParam === "true",
  });

  const { Stepper } = React.useMemo(
    () =>
      defineStepper(
        ...(appStoreApp
          ? [
              {
                id: "install",
                title: "",
              },
            ]
          : []),
        {
          id: "configure",
          title: "Configure",
        },
        ...(returnTo || pubkey ? [] : [{ id: "finalize", title: "Finalize" }])
      ),
    [appStoreApp, returnTo, pubkey]
  );

  const handleCreateApp = async (nextFunc: () => void) => {
    if (!permissions.scopes.length) {
      toast("Please specify wallet permissions.");
      return;
    }

    setLoading(true);
    try {
      const createAppRequest: CreateAppRequest = {
        name: appName,
        pubkey,
        budgetRenewal: permissions.budgetRenewal,
        maxAmount: permissions.maxAmount || 0,
        scopes: [
          ...permissions.scopes,
          ...(superuser ? ["superuser" satisfies Scope] : []),
        ] as Scope[],
        expiresAt: permissions.expiresAt?.toISOString(),
        returnTo: returnTo,
        isolated: permissions.isolated,
        metadata: {
          app_store_app_id: appStoreApp?.id,
        },
        unlockPassword,
      };

      const createAppResponse = await createApp(createAppRequest);

      // dispatch a success event which can be listened to by the opener or by the app that embedded the webview
      // this gives those apps the chance to know the user has enabled the connection
      const nwcEvent = new CustomEvent("nwc:success", {
        detail: {
          relayUrl: createAppResponse.relayUrls[0], // TODO: support multiple relays
          walletPubkey: createAppResponse.walletPubkey,
          lud16: createAppResponse.lud16,
        },
      });
      window.dispatchEvent(nwcEvent);

      // notify the opener of the successful connection
      if (window.opener) {
        window.opener.postMessage(
          {
            type: "nwc:success",
            relayUrl: createAppResponse.relayUrls[0], // TODO: support multiple relays
            walletPubkey: createAppResponse.walletPubkey,
            lud16: createAppResponse.lud16,
          },
          "*"
        );
      }

      if (createAppResponse.returnTo) {
        // open connection URI directly in an app
        window.location.href = createAppResponse.returnTo;
        return;
      }
      toast("App created");
      setCreateAppResponse(createAppResponse);
      if (pubkey) {
        return;
      }

      nextFunc();
    } catch (error) {
      handleRequestError("Failed to create app", error);
    }
    setLoading(false);
  };

  if (unsupportedError) {
    return (
      <>
        <AppHeader title="Unsupported App" description={unsupportedError} />
        <p>Try the Alby Hub LDK backend for extra features.</p>
      </>
    );
  }

  return (
    <>
      <AppHeader
        title={appName ? `Connect to ${appName}` : "Connect a new app"}
        icon={
          appStoreApp?.logo ? (
            <img
              src={appStoreApp.logo}
              alt="logo"
              className="inline rounded-lg w-12 h-12"
            />
          ) : undefined
        }
        description="Configure wallet permissions for the app and follow instructions to finalize the connection"
      />

      <Stepper.Provider className="space-y-4 max-w-lg" variant="vertical">
        {({ methods }) => (
          <>
            <Stepper.Navigation>
              {methods.all.map((step) => (
                <Stepper.Step
                  key={step.id}
                  of={step.id}
                  onClick={() =>
                    methods.current.id === "configure" && step.id === "install"
                      ? methods.goTo(step.id)
                      : undefined
                  }
                >
                  <Stepper.Title>
                    {step.title ||
                      (isInstallable ? "Install" : "Open") + " " + appName}
                  </Stepper.Title>
                  {methods.when(step.id, () => (
                    <>
                      {methods.switch({
                        install: () =>
                          appStoreApp && (
                            <InstallApp appStoreApp={appStoreApp} />
                          ),
                        configure: () => (
                          <div className="flex flex-col gap-4">
                            <SuperuserConfirmPasswordDialog
                              open={showSuperuserConfirmPasswordDialog}
                              setOpen={setShowSuperuserConfirmPasswordDialog}
                              onSubmit={() => {
                                handleCreateApp(methods.next);
                              }}
                              unlockPassword={unlockPassword}
                              setUnlockPassword={setUnlockPassword}
                            />
                            {!appStoreApp && (
                              <div className="w-full grid gap-1.5">
                                <Label htmlFor="name">Name</Label>
                                <Input
                                  autoFocus
                                  type="text"
                                  name="name"
                                  value={appName}
                                  id="name"
                                  onChange={(e) => setAppName(e.target.value)}
                                  required
                                  autoComplete="off"
                                />
                                <p className="text-xs text-muted-foreground">
                                  Name of the app or purpose of the connection
                                </p>
                              </div>
                            )}
                            <div className="flex flex-col gap-2 w-full">
                              <Permissions
                                capabilities={capabilities}
                                permissions={permissions}
                                setPermissions={setPermissions}
                                isNewConnection
                                scopesReadOnly={
                                  !!reqMethodsParam ||
                                  !!notificationTypesParam ||
                                  !!isolatedParam
                                }
                                budgetReadOnly={!!budgetMaxAmountMsatParam}
                                expiresAtReadOnly={!!expiresAtParam}
                              />
                            </div>
                            {appStoreApp?.superuser && (
                              <div className="flex mt-2">
                                <Checkbox
                                  id="superuser"
                                  required
                                  checked={superuser}
                                  onCheckedChange={() =>
                                    setSuperuser(!superuser)
                                  }
                                  className="mt-0.5"
                                />
                                <Label
                                  htmlFor="superuser"
                                  className="ml-2 text-sm text-foreground flex flex-col items-start justify-center"
                                >
                                  <div>
                                    Enable accepting connections to other apps
                                  </div>
                                  <div className="text-muted-foreground font-normal">
                                    Allow this app to let you authorize new
                                    connections to your Alby Hub.
                                  </div>
                                </Label>
                              </div>
                            )}

                            {returnTo && (
                              <p className="text-xs text-muted-foreground">
                                You will automatically return to {returnTo}
                              </p>
                            )}
                          </div>
                        ),
                        finalize: () =>
                          createAppResponse && (
                            <div className="pl-8 max-w-md">
                              <FinalizeConnection
                                createAppResponse={createAppResponse}
                                appStoreApp={appStoreApp}
                              />
                            </div>
                          ),
                      })}
                      {(!methods.isLast || returnTo || pubkey) && (
                        <Stepper.Controls className="mt-6">
                          {!methods.isFirst && (
                            <Button
                              type="button"
                              variant="secondary"
                              onClick={methods.prev}
                            >
                              Back
                            </Button>
                          )}
                          <LoadingButton
                            loading={isLoading}
                            onClick={
                              step.id === "configure"
                                ? () =>
                                    superuser
                                      ? setShowSuperuserConfirmPasswordDialog(
                                          true
                                        )
                                      : handleCreateApp(methods.next)
                                : methods.next
                            }
                          >
                            {step.id === "configure" && pubkey
                              ? "Connect"
                              : "Next"}
                          </LoadingButton>
                        </Stepper.Controls>
                      )}
                    </>
                  ))}
                </Stepper.Step>
              ))}
            </Stepper.Navigation>
          </>
        )}
      </Stepper.Provider>
    </>
  );
};

export default NewApp;

function FinalizeConnection({
  createAppResponse,
  appStoreApp,
}: {
  createAppResponse: CreateAppResponse;
  appStoreApp: AppStoreApp | undefined;
}) {
  const navigate = useNavigate();

  const pairingUri = createAppResponse.pairingUri;
  const { data: app } = useApp(createAppResponse.id, true);

  React.useEffect(() => {
    if (app?.lastUsedAt) {
      toast("Connection established!", {
        description: "You can now use the app with your Alby Hub.",
      });
      navigate("/apps?tab=connected-apps");
    }
  }, [app?.lastUsedAt, navigate]);

  if (!createAppResponse) {
    return <Navigate to="/apps/new" />;
  }

  return (
    <>
      <div className="flex flex-col gap-3 sensitive">
        {appStoreApp ? (
          <>{appStoreApp.finalizeGuide}</>
        ) : (
          <ol className="list-decimal list-inside">
            <li>Open the app you wish to connect to</li>
            <li>
              Find settings to connect your wallet (may be under{" "}
              <span className="font-semibold">Nostr Wallet Connect</span> or{" "}
              <span className="font-semibold">NWC</span>).
            </li>
            <li>Scan or paste the connection secret</li>
          </ol>
        )}

        {app?.isolated && (
          <li>
            Optional: Top up sub-wallet balance (
            {new Intl.NumberFormat().format(Math.floor(app.balance / 1000))}{" "}
            sats){" "}
            <IsolatedAppTopupDialog appId={app.id}>
              <Button size="sm" variant="secondary">
                Top Up
              </Button>
            </IsolatedAppTopupDialog>
          </li>
        )}
        {app && (
          <ConnectAppCard
            app={app}
            pairingUri={pairingUri}
            appStoreApp={appStoreApp}
          />
        )}
      </div>
    </>
  );
}

type SuperuserConfirmPasswordDialogProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  onSubmit: () => void;
  unlockPassword: string;
  setUnlockPassword: (password: string) => void;
};

function SuperuserConfirmPasswordDialog({
  open,
  setOpen,
  onSubmit,
  unlockPassword,
  setUnlockPassword,
}: SuperuserConfirmPasswordDialogProps) {
  return (
    <AlertDialog open={open}>
      <AlertDialogContent>
        <form
          onSubmit={(e: React.FormEvent) => {
            e.preventDefault();
            onSubmit();
          }}
        >
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm New Connection</AlertDialogTitle>
            <AlertDialogDescription>
              <div className="flex flex-col">
                <p>
                  Alby Go will be given permission to create other app
                  connections which can spend your balance.
                </p>

                <p className="mt-4">
                  Warning: Alby Go can create connections with a larger budget
                  than the one set for Alby Go. Make sure to always set a
                  budget.
                </p>

                <p className="mt-4">
                  Please enter your unlock password to continue.
                </p>
                <div className="grid gap-1.5 mt-4">
                  <Label htmlFor="password">Unlock Password</Label>
                  <PasswordInput
                    id="password"
                    onChange={setUnlockPassword}
                    autoFocus
                    value={unlockPassword}
                  />
                </div>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter className="mt-3">
            <AlertDialogCancel onClick={() => setOpen(false)}>
              Cancel
            </AlertDialogCancel>
            <Button type="submit">Confirm</Button>
          </AlertDialogFooter>
        </form>
      </AlertDialogContent>
    </AlertDialog>
  );
}
