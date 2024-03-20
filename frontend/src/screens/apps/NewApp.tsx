import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import toast from "react-hot-toast";

import {
  BudgetRenewalType,
  CreateAppResponse,
  RequestMethodType,
  nip47MethodDescriptions,
  validBudgetRenewals,
} from "src/types";
import { useCSRF } from "src/hooks/useCSRF";
import { EditIcon } from "src/components/icons/EditIcon";

import { request } from "src/utils/request"; // build the project for this to appear
import { handleRequestError } from "src/utils/handleRequestError";
import Input from "src/components/Input";
import Permissions from "../../components/Permissions";

const NewApp = () => {
  const { data: csrf } = useCSRF();
  const navigate = useNavigate();

  const location = useLocation();
  const queryParams = new URLSearchParams(location.search);

  const [isEditing, setEditing] = useState(false);

  const nameParam = (queryParams.get("name") || queryParams.get("c")) ?? "";
  const [appName, setAppName] = useState(nameParam);
  const pubkey = queryParams.get("pubkey") ?? "";
  const returnTo = queryParams.get("return_to") ?? "";

  const budgetRenewalParam = queryParams.get(
    "budget_renewal"
  ) as BudgetRenewalType;
  const reqMethodsParam = queryParams.get("request_methods") ?? "";
  const maxAmountParam = queryParams.get("max_amount") ?? "";
  const expiresAtParam = queryParams.get("expires_at") ?? "";

  const parseRequestMethods = (reqParam: string): Set<RequestMethodType> => {
    const methods = reqParam
      ? reqParam.split(" ")
      : Object.keys(nip47MethodDescriptions);
    // Create a Set of RequestMethodType from the array
    const requestMethodsSet = new Set<RequestMethodType>(
      methods as RequestMethodType[]
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

  const [permissions, setPermissions] = useState({
    requestMethods: parseRequestMethods(reqMethodsParam),
    maxAmount: parseInt(maxAmountParam || "100000"),
    budgetRenewal: validBudgetRenewals.includes(budgetRenewalParam)
      ? budgetRenewalParam
      : "monthly",
    expiresAt: parseExpiresParam(expiresAtParam),
  });

  const handlePermissionsChange = (
    changedPermissions: Partial<typeof permissions>
  ) => {
    setPermissions((prev) => ({ ...prev, ...changedPermissions }));
  };

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
      navigate("/apps/created", {
        state: createAppResponse,
      });
      toast.success("App created!");
    } catch (error) {
      handleRequestError("Failed to create app", error);
    }
  };

  return (
    <div className="container max-w-screen-lg mt-6">
      <form onSubmit={handleSubmit} acceptCharset="UTF-8">
        <div className="bg-white dark:bg-surface-02dp rounded-md shadow p-4 md:p-8">
          <h2 className="font-bold text-2xl font-headline mb-4 dark:text-white">
            {nameParam ? `Connect to ${appName}` : "Connect a new app"}
          </h2>
          {!nameParam && (
            <>
              <label
                htmlFor="name"
                className="block font-medium text-gray-900 dark:text-white"
              >
                Name
              </label>
              <Input
                readOnly={!!nameParam}
                type="text"
                name="name"
                value={appName}
                id="name"
                onChange={(e) => setAppName(e.target.value)}
                required
                autoComplete="off"
              />
              <p className="mt-1 mb-6 text-xs text-gray-500 dark:text-gray-400">
                Name of the app or purpose of the connection
              </p>
            </>
          )}
          <div className="flex justify-between items-center mb-2 text-gray-800 dark:text-white">
            <p className="text-lg font-medium">Authorize the app to:</p>
            {!reqMethodsParam && !isEditing && (
              <EditIcon
                onClick={() => setEditing(true)}
                className="text-gray-800 dark:text-gray-300 cursor-pointer w-6"
              />
            )}
          </div>

          <Permissions
            initialPermissions={permissions}
            onPermissionsChange={handlePermissionsChange}
            isEditing={isEditing}
            isNew
          />
        </div>

        <div className="mt-6 flex flex-col sm:flex-row sm:justify-center px-4 md:px-8">
          {!pubkey && (
            <Link
              to="/apps"
              className="inline-flex p-4 underline cursor-pointer duration-150 items-center justify-center text-gray-700 dark:text-neutral-300 w-full sm:w-[250px] order-last sm:order-first"
            >
              Cancel
            </Link>
          )}
          <button
            type="submit"
            className="inline-flex w-full sm:w-[250px] bg-indigo-500 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium hover:bg-indigo-700 items-center justify-center px-5 py-3 rounded-md shadow text-white transition"
          >
            {pubkey ? "Connect" : "Next"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default NewApp;
