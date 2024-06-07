import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

import { useCSRF } from "src/hooks/useCSRF";
import {
  AppPermissions,
  BudgetRenewalType,
  CreateAppResponse,
  PermissionType,
  nip47PermissionDescriptions,
  validBudgetRenewals,
} from "src/types";

import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { Separator } from "src/components/ui/separator";
import { useToast } from "src/components/ui/use-toast";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear
import Permissions from "../../components/Permissions";
import { suggestedApps } from "../../components/SuggestedAppData";

const NewApp = () => {
  const location = useLocation();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const navigate = useNavigate();

  const queryParams = new URLSearchParams(location.search);

  const appId = queryParams.get("app") ?? "";
  const app = suggestedApps.find((app) => app.id === appId);

  const nameParam = app
    ? app.title
    : (queryParams.get("name") || queryParams.get("c")) ?? "";
  const pubkey = queryParams.get("pubkey") ?? "";
  const returnTo = queryParams.get("return_to") ?? "";

  const [appName, setAppName] = useState(nameParam);

  const budgetRenewalParam = queryParams.get(
    "budget_renewal"
  ) as BudgetRenewalType;
  const reqMethodsParam = queryParams.get("request_methods") ?? "";
  const maxAmountParam = queryParams.get("max_amount") ?? "";
  const expiresAtParam = queryParams.get("expires_at") ?? "";

  const parseRequestMethods = (reqParam: string): Set<PermissionType> => {
    const methods = reqParam
      ? reqParam.split(" ")
      : Object.keys(nip47PermissionDescriptions);
    // Create a Set of PermissionType from the array
    const requestMethodsSet = new Set<PermissionType>(
      methods as PermissionType[]
    );

    return requestMethodsSet;
  };

  const parseExpiresParam = (expiresParam: string): Date | undefined => {
    const expiresParamTimestamp = parseInt(expiresParam);
    if (!isNaN(expiresParamTimestamp)) {
      return new Date(expiresParamTimestamp * 1000);
    }
    return undefined;
  };

  const [permissions, setPermissions] = useState<AppPermissions>({
    requestMethods: parseRequestMethods(reqMethodsParam),
    maxAmount: parseInt(maxAmountParam || "100000"),
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

    try {
      const createAppResponse = await request<CreateAppResponse>("/api/apps", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: appName,
          pubkey,
          ...permissions,
          requestMethods: [...permissions.requestMethods].join(" "),
          expiresAt: permissions.expiresAt?.toISOString(),
          returnTo: returnTo,
        }),
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
          <p className="font-medium text-sm">Authorize the app to:</p>
          <Permissions
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
