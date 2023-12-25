import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useInfo } from "../../hooks/useInfo";
import {
  BudgetRenewalType,
  CreateAppResponse,
  RequestMethodType,
  User,
  UserInfo,
  nip47MethodDescriptions,
  nip47MethodIcons,
  validBudgetRenewals,
} from "../../types";

const New = () => {
  const { data: info } = useInfo();
  const navigate = useNavigate();
  const currentUser: User = (info as UserInfo).user!;

  const location = useLocation();
  const queryParams = new URLSearchParams(location.search);

  const [edit, setEdit] = useState(false);

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

  const reqMethodsParam = queryParams.get("request_methods");
  const [requestMethods, setRequestMethods] = useState(
    reqMethodsParam ?? Object.keys(nip47MethodDescriptions).join(" ")
  );

  const maxAmountParam = queryParams.get("max_amount") ?? "";
  const [maxAmount, setMaxAmount] = useState(parseInt(maxAmountParam));

  const parseExpiresParam = (expiresParam: string): string => {
    if (/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z/.test(expiresParam)) {
      const d = new Date(expiresParam);
      const isIso =
        d instanceof Date &&
        !isNaN(d.getTime()) &&
        d.toISOString() === expiresAtParam;
      if (isIso) {
        return expiresParam;
      }
    }
    if (!isNaN(parseInt(expiresParam))) {
      return new Date(parseInt(expiresAtParam as string) * 1000).toISOString();
    }
    return "";
  };

  // Only timestamp in seconds or ISO string is expected
  const expiresAtParam = parseExpiresParam(queryParams.get("expires_at") ?? "");
  const [expiresAt, setExpiresAt] = useState(expiresAtParam ?? "");
  const [days, setDays] = useState(0);
  const [expireOptions, setExpireOptions] = useState(false);

  const today = new Date();
  const handleDays = (days: number) => {
    setDays(days);
    if (!days) {
      setExpiresAt("");
      return;
    }
    const expiryDate = new Date(today.getTime() + days * 24 * 60 * 60 * 1000);
    setExpiresAt(expiryDate.toISOString());
  };

  const handleRequestMethodChange = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const rm = event.target.value;
    if (requestMethods.includes(rm)) {
      // If checked and item is already in the list, remove it
      const newMethods = requestMethods
        .split(" ")
        .filter((reqMethod) => reqMethod !== rm)
        .join(" ");
      setRequestMethods(newMethods);
    } else {
      setRequestMethods(`${requestMethods} ${rm}`);
    }
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!info) return;

    const formData = new FormData();
    formData.append("name", appName);
    formData.append("pubkey", pubkey);
    formData.append("maxAmount", maxAmount.toString());
    formData.append("budgetRenewal", budgetRenewal);
    formData.append("expiresAt", expiresAt);
    formData.append("requestMethods", requestMethods);
    formData.append("returnTo", returnTo);

    try {
      const response = await fetch("/api/apps", {
        method: "POST",
        headers: {
          "X-CSRF-Token": info.csrf,
        },
        body: formData,

        // TODO: send consider sending JSON data
        /*body: JSON.stringify({
          name: appName,
          pubkey: pubkey,
          maxAmount: maxAmount.toString(),
          budgetRenewal: budgetRenewal,
          expiresAt: expiresAt,
          requestMethods: requestMethods,
          returnTo: returnTo,
        }),*/
      });
      console.log(response);
      const createAppResponse: CreateAppResponse = await response.json();
      navigate("/apps/created", {
        state: createAppResponse,
      });
    } catch (error) {
      // TODO: Deal with invalid pubkey format error
      // Invalid expiresAt error
      console.error("Error deleting app:", error);
    }
  };

  return (
    <div>
      <h2 className="font-bold text-2xl font-headline mb-4 dark:text-white">
        {nameParam ? `Connect to ${appName}` : "Connect a new app"}
      </h2>

      <form onSubmit={handleSubmit} acceptCharset="UTF-8">
        <div className="bg-white dark:bg-surface-02dp rounded-md shadow p-4 md:p-8">
          <input type="hidden" name="_csrf" value="YOUR_CSRF_TOKEN" />
          <input type="hidden" name="pubkey" value={pubkey ?? ""} />
          <input type="hidden" name="returnTo" value={returnTo ?? ""} />

          {!nameParam ? (
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
          ) : (
            <input type="hidden" name="name" id="" value="{{.Name}}" />
          )}

          <input
            type="hidden"
            name="RequestMethods"
            id="request-methods"
            value={requestMethods}
          />
          <input
            type="hidden"
            name="ExpiresAt"
            id="expires-at"
            value={expiresAt}
          />
          <input
            type="hidden"
            name="MaxAmount"
            id="max-amount"
            value={maxAmount}
          />
          <input
            type="hidden"
            name="BudgetRenewal"
            id="budget-renewal"
            value={budgetRenewal}
          />

          <div className="flex justify-between items-center mb-2 text-gray-800 dark:text-white">
            <p className="text-lg font-medium">Authorize the app to:</p>
            {!reqMethodsParam && (
              <img
                onClick={() => setEdit(true)}
                className="dark:invert opacity-80 inline cursor-pointer w-6"
                src="/public/images/edit.svg"
              />
            )}
          </div>

          <div className="mb-6">
            <ul className="flex flex-col w-full">
              {Object.keys(nip47MethodDescriptions).map((rm, index) => (
                <li
                  key={index}
                  className={`w-full ${
                    rm == "pay_invoice" ? "order-last" : ""
                  } ${!edit && !requestMethods.includes(rm) ? "hidden" : ""}`}
                >
                  <div className="flex items-center mb-2">
                    <img
                      className={`dark:invert opacity-80 w-5 mr-3 ${
                        edit ? "hidden" : ""
                      }`}
                      src={`/public/images/${
                        nip47MethodIcons[rm as RequestMethodType]
                      }.svg`}
                    />
                    <input
                      type="checkbox"
                      id={rm}
                      value={rm}
                      checked={requestMethods.includes(rm)}
                      onChange={handleRequestMethodChange}
                      className={` ${
                        !edit ? "hidden" : ""
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
                        !requestMethods.includes(rm)
                          ? edit
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
                            <div
                              onClick={() => setMaxAmount(10000)}
                              className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                maxAmount == 10000
                                  ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                                  : "border-gray-200 dark:border-gray-400"
                              } text-center py-4 dark:text-white`}
                            >
                              10k
                              <br />
                              sats
                            </div>
                            <div
                              onClick={() => setMaxAmount(25000)}
                              className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                maxAmount == 25000
                                  ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                                  : "border-gray-200 dark:border-gray-400"
                              } text-center py-4 dark:text-white`}
                            >
                              25k
                              <br />
                              sats
                            </div>
                            <div
                              onClick={() => setMaxAmount(50000)}
                              className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                maxAmount == 50000
                                  ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                                  : "border-gray-200 dark:border-gray-400"
                              } text-center py-4 dark:text-white`}
                            >
                              50k
                              <br />
                              sats
                            </div>
                            <div
                              onClick={() => setMaxAmount(100000)}
                              className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                maxAmount == 100000
                                  ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                                  : "border-gray-200 dark:border-gray-400"
                              } text-center py-4 dark:text-white`}
                            >
                              100k
                              <br />
                              sats
                            </div>
                            <div
                              onClick={() => setMaxAmount(1000000)}
                              className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                maxAmount == 1000000
                                  ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                                  : "border-gray-200 dark:border-gray-400"
                              } text-center py-4 dark:text-white`}
                            >
                              1M
                              <br />
                              sats
                            </div>
                            <div
                              onClick={() => setMaxAmount(0)}
                              className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                maxAmount == 0
                                  ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                                  : "border-gray-200 dark:border-gray-400"
                              } text-center py-4 dark:text-white`}
                            >
                              Unlimited
                              <br />
                              #reckless
                            </div>
                          </div>
                        </>
                      ) : (
                        <>
                          <p className="text-gray-600 dark:text-gray-300 text-sm">
                            <span className="capitalize">{budgetRenewal}</span>{" "}
                            budget:
                            {maxAmount} sats
                          </p>
                        </>
                      )}
                    </div>
                  )}
                </li>
              ))}
            </ul>
          </div>

          {!expiresAtParam ? (
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
                    <div
                      onClick={() => handleDays(7)}
                      className={`cursor-pointer rounded border-2 ${
                        days == 7
                          ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                          : "border-gray-200 dark:border-gray-400"
                      } text-center py-4`}
                    >
                      1 week
                    </div>
                    <div
                      onClick={() => handleDays(30)}
                      className={`cursor-pointer rounded border-2 ${
                        days == 30
                          ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                          : "border-gray-200 dark:border-gray-400"
                      } text-center py-4`}
                    >
                      1 month
                    </div>
                    <div
                      onClick={() => handleDays(365)}
                      className={`cursor-pointer rounded border-2 ${
                        days == 365
                          ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                          : "border-gray-200 dark:border-gray-400"
                      } text-center py-4`}
                    >
                      1 year
                    </div>
                    <div
                      onClick={() => handleDays(0)}
                      className={`cursor-pointer rounded border-2 ${
                        days == 0
                          ? "border-purple-700 dark:border-purple-300 text-purple-700 bg-purple-100 dark:bg-purple-900"
                          : "border-gray-200 dark:border-gray-400"
                      } text-center py-4`}
                    >
                      Never
                    </div>
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
                {new Date(parseInt(expiresAtParam) * 1000).toISOString()}
              </p>
            </>
          )}

          {currentUser.email && (
            <p className="mt-8 pt-4 border-t border-gray-300 dark:border-gray-700 text-sm text-gray-500 dark:text-neutral-300 text-center">
              You're logged in as{" "}
              <span className="font-mono">{currentUser.email}</span>
              <br />
            </p>
          )}

          <div className="mt-6 flex flex-col sm:flex-row sm:justify-center">
            <button
              type="submit"
              className="inline-flex w-full sm:w-[250px] bg-purple-700 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium hover:bg-purple-900 items-center justify-center px-5 py-3 rounded-md shadow text-white transition"
            >
              {pubkey ? "Connect" : "Next"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
};

export default New;
