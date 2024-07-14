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
  Scope,
  WalletCapabilities,
  scopeDescriptions,
  scopeIconMap,
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

  const [isScopesEditable, setScopesEditable] = React.useState(
    isNewConnection ? !initialPermissions.scopes.size : canEditPermissions
  );
  const [isBudgetAmountEditable, setBudgetAmountEditable] = React.useState(
    isNewConnection
      ? Number.isNaN(initialPermissions.maxAmount)
      : canEditPermissions
  );
  const [isExpiryEditable, setExpiryEditable] = React.useState(
    isNewConnection ? !initialPermissions.expiresAt : canEditPermissions
  );

  // triggered when changes are saved in show app
  React.useEffect(() => {
    if (isNewConnection || canEditPermissions) {
      return;
    }
    setPermissions(initialPermissions);
  }, [canEditPermissions, isNewConnection, initialPermissions]);

  // triggered when edit mode is toggled in show app
  React.useEffect(() => {
    if (isNewConnection) {
      return;
    }
    setScopesEditable(canEditPermissions);
    setBudgetAmountEditable(canEditPermissions);
    setExpiryEditable(canEditPermissions);
  }, [canEditPermissions, isNewConnection]);

  const [showBudgetOptions, setShowBudgetOptions] = React.useState(
    permissions.scopes.has("pay_invoice")
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

  const handleExpiryChange = React.useCallback(
    (expiryDate?: Date) => {
      handlePermissionsChange({ expiresAt: expiryDate });
    },
    [handlePermissionsChange]
  );

  return (
    <div className="max-w-lg">
      {isScopesEditable ? (
        <Scopes
          capabilities={capabilities}
          scopes={permissions.scopes}
          onScopesChanged={handleScopeChange}
        />
      ) : (
        <>
          <p className="text-sm font-medium mb-2">Scopes</p>
          <div className="flex flex-col mb-2">
            {[...permissions.scopes].map((scope) => {
              const PermissionIcon = scopeIconMap[scope];
              return (
                <div
                  key={scope}
                  className={cn(
                    "flex items-center mb-2",
                    scope == "pay_invoice" && "order-last"
                  )}
                >
                  <PermissionIcon className="mr-2 w-4 h-4" />
                  <p className="text-sm">{scopeDescriptions[scope]}</p>
                </div>
              );
            })}
          </div>
        </>
      )}
      {permissions.scopes.has("pay_invoice") &&
        (!isBudgetAmountEditable ? (
          <div className="pl-4 ml-2 border-l-2 border-l-primary mb-4">
            <div className="flex flex-col gap-2 text-muted-foreground text-sm">
              <p className="capitalize">
                <span className="text-primary-foreground font-medium">
                  Budget Renewal:
                </span>{" "}
                {permissions.budgetRenewal || "Never"}
              </p>
              <p>
                <span className="text-primary-foreground font-medium">
                  Budget Amount:
                </span>{" "}
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
                  handleBudgetMaxAmountChange(100000);
                  setShowBudgetOptions(true);
                }}
                className={cn(
                  "mr-4",
                  (!isExpiryEditable || showExpiryOptions) && "mb-4"
                )}
              >
                <PlusCircle className="w-4 h-4 mr-2" />
                Set budget
              </Button>
            )}
            {showBudgetOptions && (
              <>
                <BudgetRenewalSelect
                  value={permissions.budgetRenewal}
                  onChange={handleBudgetRenewalChange}
                />
                <BudgetAmountSelect
                  value={permissions.maxAmount}
                  onChange={handleBudgetMaxAmountChange}
                />
              </>
            )}
          </>
        ))}

      {!isExpiryEditable ? (
        <>
          <p className="text-sm font-medium mb-2">Connection expiry</p>
          <p className="text-muted-foreground text-sm">
            {permissions.expiresAt &&
            new Date(permissions.expiresAt).getFullYear() !== 1
              ? new Date(permissions.expiresAt).toString()
              : "This app will never expire"}
          </p>
        </>
      ) : (
        <>
          {!showExpiryOptions && (
            <Button
              type="button"
              variant="secondary"
              onClick={() => setShowExpiryOptions(true)}
            >
              <PlusCircle className="w-4 h-4 mr-2" />
              Set expiration time
            </Button>
          )}

          {showExpiryOptions && (
            <ExpirySelect
              value={permissions.expiresAt}
              onChange={handleExpiryChange}
            />
          )}
        </>
      )}
    </div>
  );
};

export default Permissions;
