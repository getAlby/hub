import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import toast from "react-hot-toast";

import {
  BudgetRenewalType,
  CreateAppResponse,
  IconMap,
  NIP_47_GET_BALANCE_METHOD,
  NIP_47_GET_INFO_METHOD,
  NIP_47_LIST_TRANSACTIONS_METHOD,
  NIP_47_LOOKUP_INVOICE_METHOD,
  NIP_47_MAKE_INVOICE_METHOD,
  NIP_47_PAY_INVOICE_METHOD,
  RequestMethodType,
  nip47MethodDescriptions,
  validBudgetRenewals,
} from "src/types";
import { useCSRF } from "src/hooks/useCSRF";
import { EditIcon } from "src/components/icons/EditIcon";
import { WalletIcon } from "src/components/icons/WalletIcon";
import { LightningIcon } from "src/components/icons/LightningIcon";
import { InvoiceIcon } from "src/components/icons/InvoiceIcon";
import { SearchIcon } from "src/components/icons/SearchIcon";
import { TransactionsIcon } from "src/components/icons/TransactionsIcon";
import { request } from "src/utils/request"; // build the project for this to appear
import { handleRequestError } from "src/utils/handleRequestError";

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
  const budgetRenewal: BudgetRenewalType = validBudgetRenewals.includes(
    budgetRenewalParam
  )
    ? budgetRenewalParam
    : "monthly";

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

  const reqMethodsParam = queryParams.get("request_methods") ?? "";
  const [requestMethods, setRequestMethods] = useState(
    parseRequestMethods(reqMethodsParam)
  );

  const maxAmountParam = queryParams.get("max_amount") ?? "";
  const [maxAmount, setMaxAmount] = useState(
    parseInt(maxAmountParam || "100000")
  );

  const parseExpiresParam = (expiresParam: string): Date | undefined => {
    const expiresParamTimestamp = parseInt(expiresParam);
    if (!isNaN(expiresParamTimestamp)) {
      return new Date(expiresParamTimestamp * 1000);
    }
    return undefined;
  };

  const [expiresAt, setExpiresAt] = useState<Date | undefined>(
    parseExpiresParam(queryParams.get("expires_at") ?? "")
  );
  const [days, setDays] = useState(0);
  const [expireOptions, setExpireOptions] = useState(false);

  const today = new Date();
  const handleDays = (days: number) => {
    setDays(days);
    if (!days) {
      setExpiresAt(undefined);
      return;
    }
    const expiryDate = new Date(today.getTime() + days * 24 * 60 * 60 * 1000);
    setExpiresAt(expiryDate);
  };

  const handleRequestMethodChange = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const requestMethod = event.target.value as RequestMethodType;

    const newRequestMethods = new Set(requestMethods);

    if (newRequestMethods.has(requestMethod)) {
      // If checked and item is already in the list, remove it
      newRequestMethods.delete(requestMethod);
    } else {
      newRequestMethods.add(requestMethod);
    }

    setRequestMethods(newRequestMethods);
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
          maxAmount,
          budgetRenewal,
          expiresAt: expiresAt?.toISOString(),
          requestMethods: [...requestMethods].join(" "),
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

  const iconMap: IconMap = {
    [NIP_47_GET_BALANCE_METHOD]: WalletIcon,
    [NIP_47_GET_INFO_METHOD]: WalletIcon,
    [NIP_47_LIST_TRANSACTIONS_METHOD]: TransactionsIcon,
    [NIP_47_LOOKUP_INVOICE_METHOD]: SearchIcon,
    [NIP_47_MAKE_INVOICE_METHOD]: InvoiceIcon,
    [NIP_47_PAY_INVOICE_METHOD]: LightningIcon,
  };

  const expiryOptions: Record<string, number> = {
    "1 week": 7,
    "1 month": 30,
    "1 year": 365,
    Never: 0,
  };

  const budgetOptions: Record<string, number> = {
    "10k": 10_000,
    "25k": 25_000,
    "50k": 50_000,
    "100k": 100_000,
    "1M": 100_000_000,
    Unlimited: 0,
  };

  return (
    <div>
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
              <input
                readOnly={!!nameParam}
                type="text"
                name="name"
                value={appName}
                id="name"
                onChange={(e) => setAppName(e.target.value)}
                required
                autoComplete="off"
                className="bg-gray-50 border border-gray-300 text-gray-900 focus:ring-purple-700 dark:focus:ring-purple-600 dark:ring-offset-gray-800 focus:ring-2 text-sm rounded-lg block w-full p-2.5 dark:bg-surface-00dp dark:border-gray-700 dark:placeholder-gray-400 dark:text-white"
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

          <div className="mb-6">
            <ul className="flex flex-col w-full">
              {(
                Object.keys(nip47MethodDescriptions) as RequestMethodType[]
              ).map((rm, index) => {
                const RequestMethodIcon = iconMap[rm];
                return (
                  <li
                    key={index}
                    className={`w-full ${
                      rm == "pay_invoice" ? "order-last" : ""
                    } ${!isEditing && !requestMethods.has(rm) ? "hidden" : ""}`}
                  >
                    <div className="flex items-center mb-2">
                      {RequestMethodIcon && (
                        <RequestMethodIcon
                          className={`text-gray-800 dark:text-gray-300 w-5 mr-3 ${
                            isEditing ? "hidden" : ""
                          }`}
                        />
                      )}
                      <input
                        type="checkbox"
                        id={rm}
                        value={rm}
                        checked={requestMethods.has(rm as RequestMethodType)}
                        onChange={handleRequestMethodChange}
                        className={` ${
                          !isEditing ? "hidden" : ""
                        } w-4 h-4 mr-4 text-purple-700 bg-gray-50 border border-gray-300 rounded focus:ring-purple-700 dark:focus:ring-purple-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-surface-00dp dark:border-gray-700 cursor-pointer`}
                      />
                      <label
                        htmlFor={rm}
                        className="text-gray-800 dark:text-gray-300 cursor-pointer"
                      >
                        {nip47MethodDescriptions[rm as RequestMethodType]}
                      </label>
                    </div>
                    {rm == "pay_invoice" && (
                      <div
                        className={`pt-2 pb-2 pl-5 ml-2.5 border-l-2 border-l-gray-200 dark:border-l-gray-400 ${
                          !requestMethods.has(rm)
                            ? isEditing
                              ? "pointer-events-none opacity-30"
                              : "hidden"
                            : ""
                        }`}
                      >
                        {!maxAmountParam ? (
                          <>
                            <p className="text-gray-600 dark:text-gray-300 mb-2 text-sm">
                              Monthly budget
                            </p>
                            <div
                              id="budget-allowance-limits"
                              className="grid grid-cols-6 grid-rows-2 md:grid-rows-1 md:grid-cols-6 gap-2 text-xs text-gray-800 dark:text-neutral-200"
                            >
                              {Object.keys(budgetOptions).map((budget) => {
                                return (
                                  <div
                                    key={budget}
                                    onClick={() =>
                                      setMaxAmount(budgetOptions[budget])
                                    }
                                    className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                      maxAmount == budgetOptions[budget]
                                        ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                                        : "border-gray-200 dark:border-gray-400"
                                    } text-center py-4 dark:text-white`}
                                  >
                                    {budget}
                                    <br />
                                    {budgetOptions[budget]
                                      ? "sats"
                                      : "#reckless"}
                                  </div>
                                );
                              })}
                            </div>
                          </>
                        ) : (
                          <>
                            <p className="text-gray-600 dark:text-gray-300 text-sm">
                              <span className="capitalize">
                                {budgetRenewal}
                              </span>{" "}
                              budget: {maxAmount} sats
                            </p>
                          </>
                        )}
                      </div>
                    )}
                  </li>
                );
              })}
            </ul>
          </div>

          {!expiresAt || days ? (
            <>
              <div
                onClick={() => setExpireOptions(true)}
                className={`${
                  expireOptions ? "hidden" : ""
                } cursor-pointer text-sm font-medium text-purple-700  dark:text-purple-500`}
              >
                + Add connection expiry time
              </div>

              {expireOptions && (
                <div className="text-gray-800 dark:text-neutral-200">
                  <p className="text-lg font-medium mb-2">
                    Connection expiry time
                  </p>
                  <div
                    id="expiry-days"
                    className="grid grid-cols-4 gap-2 text-xs"
                  >
                    {Object.keys(expiryOptions).map((expiry) => {
                      return (
                        <div
                          key={expiry}
                          onClick={() => handleDays(expiryOptions[expiry])}
                          className={`cursor-pointer rounded border-2 ${
                            days == expiryOptions[expiry]
                              ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                              : "border-gray-200 dark:border-gray-400"
                          } text-center py-4`}
                        >
                          {expiry}
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </>
          ) : (
            <>
              <p className="text-lg font-medium mb-2 text-gray-800 dark:text-neutral-200">
                Connection expiry time
              </p>
              <p className="text-gray-600 dark:text-gray-300 text-sm">
                {expiresAt.toLocaleString()}
              </p>
            </>
          )}
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
            className="inline-flex w-full sm:w-[250px] bg-purple-700 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium hover:bg-purple-900 items-center justify-center px-5 py-3 rounded-md shadow text-white transition"
          >
            {pubkey ? "Connect" : "Next"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default NewApp;
