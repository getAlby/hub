import { PlusCircle } from "lucide-react";
import React from "react";
import BudgetAmountSelect from "src/components/BudgetAmountSelect";
import BudgetRenewalSelect from "src/components/BudgetRenewalSelect";
import ExpirySelect from "src/components/ExpirySelect";
import Scopes from "src/components/Scopes";
import { Button } from "src/components/ui/button";
import { cn } from "src/lib/utils";
import {
  AppPermissions,
  BudgetRenewalType,
  NIP_47_PAY_INVOICE_METHOD,
  Scope,
  WalletCapabilities,
  iconMap,
  scopeDescriptions,
} from "src/types";

interface PermissionsProps {
  capabilities: WalletCapabilities;
  initialPermissions: AppPermissions;
  onPermissionsChange: (permissions: AppPermissions) => void;
  canEditPermissions?: boolean;
  budgetUsage?: number;
  isNewConnection?: boolean;
}

const Permissions: React.FC<PermissionsProps> = ({
  capabilities,
  initialPermissions,
  onPermissionsChange,
  canEditPermissions,
  isNewConnection,
  budgetUsage,
}) => {
  const [permissions, setPermissions] = React.useState(initialPermissions);

  const [canEditBudgetAmount, setCanEditBudgetAmount] = React.useState(
    isNewConnection
      ? Number.isNaN(initialPermissions.maxAmount)
      : canEditPermissions
  );
  const [canEditExpiry, setCanEditExpiry] = React.useState(
    isNewConnection ? !initialPermissions.expiresAt : canEditPermissions
  );

  // this is triggered when changes are saved in show app
  React.useEffect(() => {
    setPermissions(initialPermissions);
  }, [initialPermissions]);

  // this is triggered when edit mode is called
  React.useEffect(() => {
    if (isNewConnection) {
      return;
    }
    setCanEditBudgetAmount(
      isNewConnection
        ? Number.isNaN(initialPermissions.maxAmount)
        : canEditPermissions
    );
    setCanEditExpiry(
      isNewConnection ? !initialPermissions.expiresAt : canEditPermissions
    );
  }, [canEditPermissions, initialPermissions, isNewConnection]);

  const [showBudgetOptions, setShowBudgetOptions] = React.useState(
    isNewConnection ? !!permissions.maxAmount : true
  );
  const [showExpiryOptions, setShowExpiryOptions] = React.useState(
    isNewConnection ? !!permissions.expiresAt : true
  );

  const handlePermissionsChange = React.useCallback(
    (changedPermissions: Partial<AppPermissions>) => {
      const updatedPermissions = { ...permissions, ...changedPermissions };
      setPermissions(updatedPermissions);
      onPermissionsChange(updatedPermissions);
    },
    [permissions, onPermissionsChange]
  );

  const handleScopeChange = React.useCallback(
    (scopes: Set<Scope>) => {
      handlePermissionsChange({ scopes });
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
    (value: string) => {
      handlePermissionsChange({ budgetRenewal: value as BudgetRenewalType });
    },
    [handlePermissionsChange]
  );

  const handleExpiryDaysChange = React.useCallback(
    (expiryDays: number) => {
      if (!expiryDays) {
        handlePermissionsChange({ expiresAt: undefined });
        return;
      }
      const currentDate = new Date();
      currentDate.setDate(currentDate.getUTCDate() + expiryDays);
      currentDate.setHours(23, 59, 59);
      handlePermissionsChange({ expiresAt: currentDate });
    },
    [handlePermissionsChange]
  );

  return (
    <div className="max-w-lg">
      {canEditPermissions ? (
        <Scopes
          capabilities={capabilities}
          scopes={permissions.scopes}
          onScopeChange={handleScopeChange}
        />
      ) : (
        <>
          <p className="text-sm font-medium mb-2">Scopes</p>
          <div className="flex flex-col gap-1">
            {[...initialPermissions.scopes].map((rm) => {
              const PermissionIcon = iconMap[rm];
              return (
                <div
                  key={rm}
                  className={cn(
                    "flex items-center mb-2",
                    rm == NIP_47_PAY_INVOICE_METHOD && "order-last"
                  )}
                >
                  <PermissionIcon className="mr-2 w-4 h-4" />
                  <p className="text-sm">{scopeDescriptions[rm]}</p>
                </div>
              );
            })}
          </div>
        </>
      )}
      {capabilities.scopes.includes(NIP_47_PAY_INVOICE_METHOD) &&
        permissions.scopes.has(NIP_47_PAY_INVOICE_METHOD) &&
        (!canEditBudgetAmount ? (
          <div className="pt-2 pl-4 ml-2 border-l-2 border-l-primary">
            <div className="flex flex-col gap-2 text-muted-foreground mb-3 text-sm">
              <p className="capitalize">
                Budget Renewal: {permissions.budgetRenewal || "Never"}
              </p>
              <p>
                Budget Amount:{" "}
                {permissions.maxAmount
                  ? new Intl.NumberFormat().format(permissions.maxAmount)
                  : "∞"}
                {" sats "}
                {!isNewConnection &&
                  `(${new Intl.NumberFormat().format(budgetUsage || 0)} sats used)`}
              </p>
            </div>
          </div>
        ) : (
          <>
            {!showBudgetOptions && (
              <Button
                type="button"
                variant="secondary"
                onClick={() => {
                  setShowBudgetOptions(true);
                  handleBudgetRenewalChange("monthly");
                  handleBudgetMaxAmountChange(100000);
                }}
                className="mb-4 mr-4"
              >
                <PlusCircle className="w-4 h-4 mr-2" />
                Set budget renewal
              </Button>
            )}
            {showBudgetOptions && (
              <>
                <p className="font-medium text-sm mb-2">Budget Renewal</p>
                <div className="flex gap-2 items-center text-muted-foreground mb-4 text-sm">
                  <BudgetRenewalSelect
                    value={permissions.budgetRenewal}
                    onChange={handleBudgetRenewalChange}
                    disabled={!canEditPermissions}
                  />
                </div>
                <BudgetAmountSelect
                  value={permissions.maxAmount}
                  onChange={handleBudgetMaxAmountChange}
                />
              </>
            )}
          </>
        ))}

      {!canEditExpiry ? (
        <div className="mt-4">
          <p className="text-sm font-medium mb-2">Connection expiry</p>
          <p className="text-muted-foreground text-sm">
            {permissions.expiresAt &&
            new Date(permissions.expiresAt).getFullYear() !== 1
              ? new Date(permissions.expiresAt).toString()
              : "This app will never expire"}
          </p>
        </div>
      ) : (
        <>
          {!showExpiryOptions && (
            <Button
              type="button"
              variant="secondary"
              onClick={() => setShowExpiryOptions(true)}
              className="mb-6"
            >
              <PlusCircle className="w-4 h-4 mr-2" />
              Set expiration time
            </Button>
          )}

          {showExpiryOptions && (
            <ExpirySelect
              value={permissions.expiresAt}
              onChange={handleExpiryDaysChange}
            />
          )}
        </>
      )}
    </div>
  );
};

export default Permissions;
