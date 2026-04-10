import { AlertTriangleIcon, CoinsIcon, TimerIcon } from "lucide-react";
import React from "react";
import BudgetAmountSelect from "src/components/BudgetAmountSelect";
import BudgetRenewalSelect from "src/components/BudgetRenewalSelect";
import ExpirySelect from "src/components/ExpirySelect";
import Scopes from "src/components/Scopes";
import { Badge } from "src/components/ui/badge";
import { Switch } from "src/components/ui/switch";
import {
  DEFAULT_APP_BUDGET_RENEWAL,
  DEFAULT_APP_BUDGET_SATS,
} from "src/constants";
import { cn } from "src/lib/utils";
import {
  AppPermissions,
  BudgetRenewalType,
  Scope,
  WalletCapabilities,
  scopeDescriptions,
  scopeIconMap,
} from "src/types";

interface PermissionsProps {
  capabilities: WalletCapabilities;
  permissions: AppPermissions;
  setPermissions?: React.Dispatch<React.SetStateAction<AppPermissions>>;
  readOnly?: boolean;
}

const Permissions: React.FC<PermissionsProps> = ({
  capabilities,
  permissions,
  setPermissions,
  readOnly,
}) => {
  const [showBudgetOptions, setShowBudgetOptions] = React.useState(
    permissions.scopes.includes("pay_invoice") && permissions.maxAmount > 0
  );
  const [showExpiryOptions, setShowExpiryOptions] = React.useState(
    !!permissions.expiresAt
  );

  const handlePermissionsChange = React.useCallback(
    (changedPermissions: Partial<AppPermissions>) => {
      setPermissions?.((currentPermissions) => ({
        ...currentPermissions,
        ...changedPermissions,
      }));
    },
    [setPermissions]
  );

  const onScopesChanged = React.useCallback(
    (scopes: Scope[], isolated: boolean) => {
      handlePermissionsChange({ scopes, isolated });
    },
    [handlePermissionsChange]
  );

  const handleBudgetMaxAmountChange = React.useCallback(
    (amount: number) => {
      handlePermissionsChange({ maxAmount: amount });
    },
    [handlePermissionsChange]
  );

  const handleBudgetRenewalChange = React.useCallback(
    (budgetRenewal: BudgetRenewalType) => {
      handlePermissionsChange({ budgetRenewal });
    },
    [handlePermissionsChange]
  );

  const handleExpiryChange = React.useCallback(
    (expiryDate?: Date) => {
      handlePermissionsChange({ expiresAt: expiryDate });
    },
    [handlePermissionsChange]
  );

  return (
    <div className={cn("space-y-4", !readOnly && "max-w-lg")}>
      {!readOnly ? (
        <Scopes
          capabilities={capabilities}
          scopes={permissions.scopes}
          isolated={permissions.isolated}
          onScopesChanged={onScopesChanged}
        />
      ) : (
        <>
          <p className="text-sm font-medium mb-2">This app is authorized to:</p>
          <div className="flex flex-wrap gap-2 mb-4">
            {[...permissions.scopes].map((scope) => {
              const PermissionIcon = scopeIconMap[scope];
              return (
                <Badge
                  variant="secondary"
                  key={scope}
                  className={cn(
                    "flex items-center font-normal py-1 rounded-full px-3"
                  )}
                >
                  <PermissionIcon className="mr-1 size-4" />
                  <p className="text-sm">{scopeDescriptions[scope]}</p>
                </Badge>
              );
            })}
          </div>
        </>
      )}

      {/* We skip read only component here as budget is shown in AppUsage */}
      {permissions.scopes.includes("pay_invoice") && !readOnly && (
        <div className="rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <label
              htmlFor="budget-toggle"
              className="flex items-center gap-2 cursor-pointer"
            >
              <CoinsIcon className="size-5 text-muted-foreground" />
              <div className="flex flex-col gap-0.5">
                <p className="text-sm font-medium">Budget</p>
                <p className="text-xs text-muted-foreground">
                  Limit how much this app can spend
                </p>
              </div>
            </label>
            <Switch
              id="budget-toggle"
              checked={showBudgetOptions}
              onCheckedChange={(checked) => {
                if (checked) {
                  handleBudgetRenewalChange(DEFAULT_APP_BUDGET_RENEWAL);
                  handleBudgetMaxAmountChange(DEFAULT_APP_BUDGET_SATS);
                } else {
                  handleBudgetRenewalChange("never");
                  handleBudgetMaxAmountChange(0);
                }
                setShowBudgetOptions(checked);
              }}
            />
          </div>
          {showBudgetOptions && (
            <div className="mt-4">
              <BudgetRenewalSelect
                value={permissions.budgetRenewal}
                onChange={handleBudgetRenewalChange}
              />
              <BudgetAmountSelect
                value={permissions.maxAmount}
                onChange={handleBudgetMaxAmountChange}
              />
            </div>
          )}
        </div>
      )}

      {!readOnly ? (
        <div className="rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <label
              htmlFor="expiry-toggle"
              className="flex items-center gap-2 cursor-pointer"
            >
              <TimerIcon className="size-5 text-muted-foreground" />
              <div className="flex flex-col gap-0.5">
                <p className="text-sm font-medium">Expiration</p>
                <p className="text-xs text-muted-foreground">
                  Automatically expire this connection
                </p>
              </div>
            </label>
            <Switch
              id="expiry-toggle"
              checked={showExpiryOptions}
              onCheckedChange={(checked) => {
                if (checked) {
                  const defaultExpiry = new Date();
                  defaultExpiry.setFullYear(defaultExpiry.getFullYear() + 1);
                  defaultExpiry.setHours(23, 59, 59);
                  handleExpiryChange(defaultExpiry);
                } else {
                  handleExpiryChange(undefined);
                }
                setShowExpiryOptions(checked);
              }}
            />
          </div>
          {showExpiryOptions && (
            <div className="mt-4">
              <ExpirySelect
                value={permissions.expiresAt}
                onChange={handleExpiryChange}
              />
            </div>
          )}
        </div>
      ) : (
        <div>
          <p className="text-sm font-medium mb-2">Connection expiry</p>
          <p className="text-muted-foreground text-sm">
            {permissions.expiresAt
              ? new Date(permissions.expiresAt).toString()
              : "This app will never expire"}
          </p>
        </div>
      )}

      {permissions.scopes.includes("superuser") && (
        <>
          <div className="flex items-center gap-2 mt-4">
            <AlertTriangleIcon className="size-4" />
            <p className="text-sm font-medium">
              This app can create other app connections
            </p>
          </div>
          <p className="text-muted-foreground text-sm">
            Make sure to set budgets on connections created by this app.
          </p>
        </>
      )}
    </div>
  );
};

export default Permissions;
