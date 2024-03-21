import React, { useEffect, useState } from "react";
import {
  AppPermissions,
  RequestMethodType,
  budgetOptions,
  expiryOptions,
  nip47MethodDescriptions,
  iconMap,
  BudgetRenewalType,
  validBudgetRenewals,
} from "src/types";

interface PermissionsProps {
  initialPermissions: AppPermissions;
  onPermissionsChange: (permissions: AppPermissions) => void;
  budgetUsage?: number;
  isEditing: boolean;
  isNew?: boolean;
}

const Permissions: React.FC<PermissionsProps> = ({
  initialPermissions,
  onPermissionsChange,
  isEditing,
  isNew,
  budgetUsage,
}) => {
  const [permissions, setPermissions] = React.useState(initialPermissions);

  const [days, setDays] = useState(isNew ? 0 : -1);
  const [expireOptions, setExpireOptions] = useState(!isNew);

  useEffect(() => {
    setPermissions(initialPermissions);
  }, [initialPermissions]);

  const handlePermissionsChange = (
    changedPermissions: Partial<AppPermissions>
  ) => {
    const updatedPermissions = { ...permissions, ...changedPermissions };
    setPermissions(updatedPermissions);
    onPermissionsChange(updatedPermissions);
  };

  const handleRequestMethodChange = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    if (!isEditing) {
      return;
    }
    const requestMethod = event.target.value as RequestMethodType;
    const newRequestMethods = new Set(permissions.requestMethods);
    if (newRequestMethods.has(requestMethod)) {
      newRequestMethods.delete(requestMethod);
    } else {
      newRequestMethods.add(requestMethod);
    }
    handlePermissionsChange({ requestMethods: newRequestMethods });
  };

  const handleMaxAmountChange = (amount: number) => {
    handlePermissionsChange({ maxAmount: amount });
  };

  const handleBudgetRenewalChange = (
    event: React.ChangeEvent<HTMLSelectElement>
  ) => {
    const budgetRenewal = event.target.value as BudgetRenewalType;
    handlePermissionsChange({ budgetRenewal });
  };

  const handleDaysChange = (days: number) => {
    setDays(days);
    if (!days) {
      handlePermissionsChange({ expiresAt: undefined });
      return;
    }
    const currentDate = new Date();
    const expiryDate = new Date(
      Date.UTC(
        currentDate.getUTCFullYear(),
        currentDate.getUTCMonth(),
        currentDate.getUTCDate() + days,
        23,
        59,
        59,
        0
      )
    );
    handlePermissionsChange({ expiresAt: expiryDate });
  };

  return (
    <div>
      <div className="mb-6">
        <ul className="flex flex-col w-full">
          {(Object.keys(nip47MethodDescriptions) as RequestMethodType[]).map(
            (rm, index) => {
              const RequestMethodIcon = iconMap[rm];
              return (
                <li
                  key={index}
                  className={`w-full ${
                    rm == "pay_invoice" ? "order-last" : ""
                  } ${
                    !isEditing && !permissions.requestMethods.has(rm)
                      ? "hidden"
                      : ""
                  }`}
                >
                  <div className="flex items-center mb-2">
                    {RequestMethodIcon && (
                      <RequestMethodIcon
                        className={`text-gray-800 dark:text-gray-300 w-4 mr-3 ${
                          isEditing ? "hidden" : ""
                        }`}
                      />
                    )}
                    <input
                      type="checkbox"
                      id={rm}
                      value={rm}
                      checked={permissions.requestMethods.has(rm)}
                      onChange={handleRequestMethodChange}
                      className={`${
                        !isEditing ? "hidden" : ""
                      } w-4 h-4 mr-4 text-indigo-500 bg-gray-50 border border-gray-300 rounded focus:ring-indigo-500 dark:focus:ring-indigo-400 dark:ring-offset-gray-800 focus:ring-2 dark:bg-surface-00dp dark:border-gray-700 cursor-pointer`}
                    />
                    <label
                      htmlFor={rm}
                      className={`text-gray-800 dark:text-gray-300 ${
                        isEditing && "cursor-pointer"
                      }`}
                    >
                      {nip47MethodDescriptions[rm]}
                    </label>
                  </div>
                  {rm == "pay_invoice" && (
                    <div
                      className={`pt-2 pb-2 pl-5 ml-2.5 border-l-2 border-l-gray-200 dark:border-l-gray-400 ${
                        !permissions.requestMethods.has(rm)
                          ? isEditing
                            ? "pointer-events-none opacity-30"
                            : "hidden"
                          : ""
                      }`}
                    >
                      {isEditing ? (
                        <>
                          <p className="text-gray-600 dark:text-gray-300 mb-3 text-sm capitalize">
                            Budget Renewal:
                            {!isEditing ? (
                              permissions.budgetRenewal
                            ) : (
                              <select
                                id="budgetRenewal"
                                value={permissions.budgetRenewal}
                                onChange={handleBudgetRenewalChange}
                                className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 ml-2 p-2.5 pr-10 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-400 dark:focus:border-indigo-400"
                                disabled={!isEditing}
                              >
                                {validBudgetRenewals.map((renewalOption) => (
                                  <option
                                    key={renewalOption || "never"}
                                    value={renewalOption || "never"}
                                  >
                                    {renewalOption
                                      ? renewalOption.charAt(0).toUpperCase() +
                                        renewalOption.slice(1)
                                      : "Never"}
                                  </option>
                                ))}
                              </select>
                            )}
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
                                    handleMaxAmountChange(budgetOptions[budget])
                                  }
                                  className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                    permissions.maxAmount ==
                                    budgetOptions[budget]
                                      ? "border-indigo-500 dark:border-indigo-400 text-indigo-500 bg-indigo-100 dark:bg-indigo-900"
                                      : "border-gray-200 dark:border-gray-400"
                                  } text-center py-4 dark:text-white`}
                                >
                                  {budget}
                                  <br />
                                  {budgetOptions[budget] ? "sats" : "#reckless"}
                                </div>
                              );
                            })}
                          </div>
                        </>
                      ) : isNew ? (
                        <>
                          <p className="text-gray-600 dark:text-gray-300 text-sm">
                            <span className="capitalize">
                              {permissions.budgetRenewal}
                            </span>{" "}
                            budget: {permissions.maxAmount} sats
                          </p>
                        </>
                      ) : (
                        <table className="text-gray-600 dark:text-neutral-400">
                          <tbody>
                            <tr className="text-sm">
                              <td className="pr-2">Budget Allowance:</td>
                              <td>
                                {permissions.maxAmount
                                  ? new Intl.NumberFormat().format(
                                      permissions.maxAmount
                                    )
                                  : "∞"}{" "}
                                sats (
                                {new Intl.NumberFormat().format(
                                  budgetUsage || 0
                                )}{" "}
                                sats used)
                              </td>
                            </tr>
                            <tr className="text-sm">
                              <td className="pr-2">Renews:</td>
                              <td className="capitalize">
                                {permissions.budgetRenewal || "Never"}
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      )}
                    </div>
                  )}
                </li>
              );
            }
          )}
        </ul>
      </div>

      {(isNew ? !permissions.expiresAt || days : isEditing) ? (
        <>
          <div
            onClick={() => setExpireOptions(true)}
            className={`${
              expireOptions ? "hidden" : ""
            } cursor-pointer text-sm font-medium text-indigo-500  dark:text-indigo-400`}
          >
            + Add connection expiry time
          </div>

          {expireOptions && (
            <div className="text-gray-800 dark:text-neutral-200">
              <p className="text-lg font-medium mb-2">Connection expiry time</p>
              {!isNew && (
                <p className="mb-2 text-gray-600 dark:text-gray-300 text-sm">
                  Expires:{" "}
                  {permissions.expiresAt &&
                  new Date(permissions.expiresAt).getFullYear() !== 1
                    ? new Date(permissions.expiresAt).toString()
                    : "This app will never expire"}
                </p>
              )}
              <div id="expiry-days" className="grid grid-cols-4 gap-2 text-xs">
                {Object.keys(expiryOptions).map((expiry) => {
                  return (
                    <div
                      key={expiry}
                      onClick={() => handleDaysChange(expiryOptions[expiry])}
                      className={`cursor-pointer rounded border-2 ${
                        days == expiryOptions[expiry]
                          ? "border-indigo-500 dark:border-indigo-400 text-indigo-500 bg-indigo-100 dark:bg-indigo-900"
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
            {permissions.expiresAt &&
            new Date(permissions.expiresAt).getFullYear() !== 1
              ? new Date(permissions.expiresAt).toString()
              : "This app will never expire"}
          </p>
        </>
      )}
    </div>
  );
};

export default Permissions;
