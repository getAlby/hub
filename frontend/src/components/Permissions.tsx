import { PlusCircle } from "lucide-react";
import React, { useEffect, useState } from "react";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useCapabilities } from "src/hooks/useCapabilities";
import { cn } from "src/lib/utils";
import {
  AppPermissions,
  BudgetRenewalType,
  Scope,
  budgetOptions,
  expiryOptions,
  iconMap,
  scopeDescriptions,
  validBudgetRenewals,
} from "src/types";

interface PermissionsProps {
  initialPermissions: AppPermissions;
  onPermissionsChange: (permissions: AppPermissions) => void;
  budgetUsage?: number;
  canEditPermissions: boolean;
  isNewConnection?: boolean;
}

const Permissions: React.FC<PermissionsProps> = ({
  initialPermissions,
  onPermissionsChange,
  canEditPermissions,
  isNewConnection,
  budgetUsage,
}) => {
  const [permissions, setPermissions] = React.useState(initialPermissions);
  const [days, setDays] = useState(isNewConnection ? 0 : -1);
  const [expireOptions, setExpireOptions] = useState(!isNewConnection);
  const { data: capabilities } = useCapabilities();

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

  const handleScopeChange = (scope: Scope) => {
    if (!canEditPermissions) {
      return;
    }

    let budgetRenewal = permissions.budgetRenewal;

    const newScopes = new Set(permissions.scopes);
    if (newScopes.has(scope)) {
      newScopes.delete(scope);
    } else {
      newScopes.add(scope);
      if (scope === "pay_invoice") {
        budgetRenewal = "monthly";
      }
    }

    handlePermissionsChange({
      scopes: newScopes,
      budgetRenewal,
    });
  };

  const handleMaxAmountChange = (amount: number) => {
    handlePermissionsChange({ maxAmount: amount });
  };

  const handleBudgetRenewalChange = (value: string) => {
    handlePermissionsChange({ budgetRenewal: value as BudgetRenewalType });
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
          {capabilities?.scopes.map((scope, index) => {
            const ScopeIcon = iconMap[scope];
            return (
              <li
                key={index}
                className={cn(
                  "w-full",
                  scope == "pay_invoice" ? "order-last" : "",
                  !canEditPermissions && !permissions.scopes.has(scope)
                    ? "hidden"
                    : ""
                )}
              >
                <div className="flex items-center mb-2">
                  {ScopeIcon && (
                    <ScopeIcon
                      className={cn(
                        "text-muted-foreground w-4 mr-3",
                        canEditPermissions ? "hidden" : ""
                      )}
                    />
                  )}
                  <Checkbox
                    id={scope}
                    className={cn("mr-2", !canEditPermissions ? "hidden" : "")}
                    onCheckedChange={() => handleScopeChange(scope)}
                    checked={permissions.scopes.has(scope)}
                  />
                  <Label
                    htmlFor={scope}
                    className={`${canEditPermissions && "cursor-pointer"}`}
                  >
                    {scopeDescriptions[scope]}
                  </Label>
                </div>
                {scope == "pay_invoice" && (
                  <div
                    className={cn(
                      "pt-2 pb-2 pl-5 ml-2.5 border-l-2 border-l-primary",
                      !permissions.scopes.has(scope)
                        ? canEditPermissions
                          ? "pointer-events-none opacity-30"
                          : "hidden"
                        : ""
                    )}
                  >
                    {canEditPermissions ? (
                      <>
                        <div className="flex flex-row gap-2 items-center text-muted-foreground mb-3 text-sm capitalize">
                          <p> Budget Renewal:</p>
                          {!canEditPermissions ? (
                            permissions.budgetRenewal
                          ) : (
                            <Select
                              value={permissions.budgetRenewal}
                              onValueChange={handleBudgetRenewalChange}
                              disabled={!canEditPermissions}
                            >
                              <SelectTrigger className="w-[150px]">
                                <SelectValue
                                  placeholder={permissions.budgetRenewal}
                                />
                              </SelectTrigger>
                              <SelectContent>
                                {validBudgetRenewals.map((renewalOption) => (
                                  <SelectItem
                                    key={renewalOption}
                                    value={renewalOption}
                                  >
                                    {renewalOption.charAt(0).toUpperCase() +
                                      renewalOption.slice(1)}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          )}
                        </div>
                        <div
                          id="budget-allowance-limits"
                          className="grid grid-cols-6 grid-rows-2 md:grid-rows-1 md:grid-cols-6 gap-2 text-xs"
                        >
                          {Object.keys(budgetOptions).map((budget) => {
                            return (
                              // replace with something else and then remove dark prefixes
                              <div
                                key={budget}
                                onClick={() =>
                                  handleMaxAmountChange(budgetOptions[budget])
                                }
                                className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                  permissions.maxAmount == budgetOptions[budget]
                                    ? "border-primary"
                                    : "border-muted"
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
                    ) : isNewConnection ? (
                      <>
                        <p className="text-muted-foreground text-sm">
                          <span className="capitalize">
                            {permissions.budgetRenewal}
                          </span>{" "}
                          budget: {permissions.maxAmount} sats
                        </p>
                      </>
                    ) : (
                      <table className="text-muted-foreground">
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
                              {new Intl.NumberFormat().format(budgetUsage || 0)}{" "}
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
          })}
        </ul>
      </div>

      {(
        isNewConnection ? !permissions.expiresAt || days : canEditPermissions
      ) ? (
        <>
          {!expireOptions && (
            <Button
              type="button"
              variant="secondary"
              onClick={() => setExpireOptions(true)}
            >
              <PlusCircle className="w-4 h-4 mr-2" />
              Set expiration date
            </Button>
          )}

          {expireOptions && (
            <div className="mt-5">
              <p className="font-medium text-sm mb-2">Connection expiration</p>
              {!isNewConnection && (
                <p className="mb-2 text-muted-foreground text-sm">
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
                      className={cn(
                        "cursor-pointer rounded border-2 text-center py-4",
                        days == expiryOptions[expiry]
                          ? "border-primary"
                          : "border-muted"
                      )}
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
          <p className="text-sm font-medium mb-2">Connection expiry</p>
          <p className="text-muted-foreground text-sm">
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
