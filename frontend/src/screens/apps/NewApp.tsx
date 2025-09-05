import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

import {
  AppPermissions,
  BudgetRenewalType,
  CreateAppRequest,
  Nip47NotificationType,
  Nip47RequestMethod,
  Scope,
  WalletCapabilities,
  validBudgetRenewals,
} from "src/types";

import React from "react";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { InstallApp } from "src/components/connections/InstallApp";
import { defineStepper } from "src/components/stepper";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { Separator } from "src/components/ui/separator";
import { useCapabilities } from "src/hooks/useCapabilities";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";
import Permissions from "../../components/Permissions";
import { appStoreApps } from "../../components/connections/SuggestedAppData";

const { Stepper } = defineStepper(
  {
    id: "install",
    title: "",
  },
  {
    id: "configure",
    title: "Configure",
  }
  //{ id: "finalize", title: "Finalize" }
);

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

  const navigate = useNavigate();
  const [unsupportedError, setUnsupportedError] = useState<string>();
  const [isLoading, setLoading] = React.useState(false);

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

  const handleSubmit = async (event?: React.FormEvent<HTMLFormElement>) => {
    event?.preventDefault();

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
        scopes: permissions.scopes,
        expiresAt: permissions.expiresAt?.toISOString(),
        returnTo: returnTo,
        isolated: permissions.isolated,
        metadata: {
          app_store_app_id: appStoreApp?.id,
        },
      };

      const createAppResponse = await createApp(createAppRequest);

      if (createAppResponse.returnTo) {
        // open connection URI directly in an app
        window.location.href = createAppResponse.returnTo;
        return;
      }
      navigate(`/apps/created${appStoreApp ? `?app=${appStoreApp.id}` : ""}`, {
        state: createAppResponse,
      });
      toast("App created");
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

  const permissionsComponent = (
    <div className="flex flex-col gap-2 w-full">
      <Permissions
        capabilities={capabilities}
        permissions={permissions}
        setPermissions={setPermissions}
        isNewConnection
        scopesReadOnly={
          !!reqMethodsParam || !!notificationTypesParam || !!isolatedParam
        }
        budgetReadOnly={!!budgetMaxAmountMsatParam}
        expiresAtReadOnly={!!expiresAtParam}
      />
    </div>
  );

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

      {appStoreApp && (
        <Stepper.Provider className="space-y-4 max-w-lg" variant="vertical">
          {({ methods }) => (
            <>
              <Stepper.Navigation>
                {methods.all.map((step) => (
                  <Stepper.Step
                    of={step.id}
                    onClick={() => methods.goTo(step.id)}
                  >
                    <Stepper.Title>
                      {step.title ||
                        (isInstallable ? "Install" : "Open") + " " + appName}
                    </Stepper.Title>
                    {methods.when(step.id, () => (
                      <>
                        {methods.switch({
                          install: () => (
                            <InstallApp appStoreApp={appStoreApp} />
                          ),
                          configure: () => permissionsComponent,
                        })}
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
                          <Button
                            onClick={
                              methods.isLast
                                ? () => handleSubmit(undefined)
                                : methods.next
                            }
                          >
                            Next
                          </Button>
                        </Stepper.Controls>
                      </>
                    ))}
                  </Stepper.Step>
                ))}
              </Stepper.Navigation>
            </>
          )}
        </Stepper.Provider>
      )}

      {!appStoreApp && (
        <form
          onSubmit={handleSubmit}
          acceptCharset="UTF-8"
          className="flex flex-col items-start gap-5 max-w-lg"
        >
          {!nameParam && (
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
          {permissionsComponent}

          <Separator />
          {returnTo && (
            <p className="text-xs text-muted-foreground">
              You will automatically return to {returnTo}
            </p>
          )}

          <LoadingButton loading={isLoading} type="submit">
            {pubkey ? "Connect" : "Next"}
          </LoadingButton>
        </form>
      )}
    </>
  );
};

export default NewApp;
