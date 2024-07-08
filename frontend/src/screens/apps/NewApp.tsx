import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

import { useCSRF } from "src/hooks/useCSRF";
import {
  AppPermissions,
  BudgetRenewalType,
  CreateAppRequest,
  CreateAppResponse,
  NIP_47_MAKE_INVOICE_METHOD,
  NIP_47_PAY_INVOICE_METHOD,
  Nip47NotificationType,
  Nip47RequestMethod,
  Scope,
  WalletCapabilities,
  validBudgetRenewals,
} from "src/types";

import React from "react";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { Separator } from "src/components/ui/separator";
import { useToast } from "src/components/ui/use-toast";
import { useCapabilities } from "src/hooks/useCapabilities";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear
import Permissions from "../../components/Permissions";
import { suggestedApps } from "../../components/SuggestedAppData";

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
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [unsupportedError, setUnsupportedError] = useState<string>();

  const queryParams = new URLSearchParams(location.search);

  const appId = queryParams.get("app") ?? "";
  const app = suggestedApps.find((app) => app.id === appId);

  const pubkey = queryParams.get("pubkey") ?? "";
  const returnTo = queryParams.get("return_to") ?? "";

  const nameParam = (queryParams.get("name") || queryParams.get("c")) ?? "";
  const [appName, setAppName] = useState(app ? app.title : nameParam);

  const budgetRenewalParam = queryParams.get(
    "budget_renewal"
  ) as BudgetRenewalType;
  const budgetMaxAmountParam = queryParams.get("max_amount") ?? "";
  const expiresAtParam = queryParams.get("expires_at") ?? "";

  const reqMethodsParam = queryParams.get("request_methods") ?? "";
  const notificationTypesParam = queryParams.get("notification_types") ?? "";

  const initialScopes: Set<Scope> = React.useMemo(() => {
    const methods = reqMethodsParam
      ? reqMethodsParam.split(" ")
      : // this is done for scope grouping
        capabilities.methods.includes(NIP_47_MAKE_INVOICE_METHOD)
        ? capabilities.methods.includes(NIP_47_PAY_INVOICE_METHOD)
          ? [NIP_47_MAKE_INVOICE_METHOD, NIP_47_PAY_INVOICE_METHOD]
          : [NIP_47_MAKE_INVOICE_METHOD]
        : capabilities.methods;
    const requestMethodsSet = new Set<Nip47RequestMethod>(
      methods as Nip47RequestMethod[]
    );

    const notificationTypes = notificationTypesParam
      ? notificationTypesParam.split(" ")
      : // this is done for scope grouping
        !capabilities.methods.includes(NIP_47_MAKE_INVOICE_METHOD)
        ? capabilities.notificationTypes
        : [];
    const notificationTypesSet = new Set<Nip47NotificationType>(
      notificationTypes as Nip47NotificationType[]
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

    const scopes = new Set<Scope>();
    if (
      requestMethodsSet.has("pay_keysend") ||
      requestMethodsSet.has("pay_invoice") ||
      requestMethodsSet.has("multi_pay_invoice") ||
      requestMethodsSet.has("multi_pay_keysend")
    ) {
      scopes.add("pay_invoice");
    }

    if (requestMethodsSet.has("get_info")) {
      scopes.add("get_info");
    }
    if (requestMethodsSet.has("get_balance")) {
      scopes.add("get_balance");
    }
    if (requestMethodsSet.has("make_invoice")) {
      scopes.add("make_invoice");
    }
    if (requestMethodsSet.has("lookup_invoice")) {
      scopes.add("lookup_invoice");
    }
    if (requestMethodsSet.has("list_transactions")) {
      scopes.add("list_transactions");
    }
    if (requestMethodsSet.has("sign_message")) {
      scopes.add("sign_message");
    }
    if (notificationTypes.length) {
      scopes.add("notifications");
    }

    return scopes;
  }, [
    capabilities.methods,
    capabilities.notificationTypes,
    notificationTypesParam,
    reqMethodsParam,
  ]);

  const parseExpiresParam = (expiresParam: string): Date | undefined => {
    const expiresParamTimestamp = parseInt(expiresParam);
    if (!isNaN(expiresParamTimestamp)) {
      return new Date(expiresParamTimestamp * 1000);
    }
    return undefined;
  };

  const [permissions, setPermissions] = useState<AppPermissions>({
    scopes: initialScopes,
    maxAmount: parseInt(budgetMaxAmountParam),
    budgetRenewal: validBudgetRenewals.includes(budgetRenewalParam)
      ? budgetRenewalParam
      : "monthly",
    expiresAt: parseExpiresParam(expiresAtParam),
  });

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!csrf) {
      throw new Error("No CSRF token");
    }

    if (!permissions.scopes.size) {
      toast({ title: "Please specify wallet permissions." });
      return;
    }

    try {
      const createAppRequest: CreateAppRequest = {
        name: appName,
        pubkey,
        budgetRenewal: permissions.budgetRenewal,
        maxAmount: permissions.maxAmount,
        scopes: Array.from(permissions.scopes),
        expiresAt: permissions.expiresAt?.toISOString(),
        returnTo: returnTo,
      };

      const createAppResponse = await request<CreateAppResponse>("/api/apps", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(createAppRequest),
      });

      if (!createAppResponse) {
        throw new Error("no create app response received");
      }

      if (createAppResponse.returnTo) {
        // open connection URI directly in an app
        window.location.href = createAppResponse.returnTo;
        return;
      }
      navigate(`/apps/created${app ? `?app=${app.id}` : ""}`, {
        state: createAppResponse,
      });
      toast({ title: "App created" });
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    }
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
        title={nameParam ? `Connect to ${appName}` : "Connect a new app"}
        description="Configure wallet permissions for the app and follow instructions to finalise the connection"
      />
      <form
        onSubmit={handleSubmit}
        acceptCharset="UTF-8"
        className="flex flex-col items-start gap-5 max-w-lg"
      >
        {app && (
          <div className="flex flex-row items-center gap-3">
            <img src={app.logo} className="h-12 w-12" />
            <h2 className="font-semibold text-lg">{app.title}</h2>
          </div>
        )}
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
        <div className="flex flex-col gap-2 w-full">
          <Permissions
            capabilities={capabilities}
            initialPermissions={permissions}
            onPermissionsChange={setPermissions}
            canEditPermissions={!reqMethodsParam}
            isNewConnection
          />
        </div>

        <Separator />
        {returnTo && (
          <p className="text-xs text-muted-foreground">
            You will automatically return to {returnTo}
          </p>
        )}

        <Button type="submit">{pubkey ? "Connect" : "Next"}</Button>
      </form>
    </>
  );
};

export default NewApp;
