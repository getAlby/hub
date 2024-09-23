import { PlusCircle } from "lucide-react";
import React from "react";
import BudgetAmountSelect from "src/components/BudgetAmountSelect";
import BudgetRenewalSelect from "src/components/BudgetRenewalSelect";
import ExpirySelect from "src/components/ExpirySelect";
import Scopes from "src/components/Scopes";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Switch } from "src/components/ui/switch";
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
  setPermissions: React.Dispatch<React.SetStateAction<AppPermissions>>;
  readOnly?: boolean;
  scopesReadOnly?: boolean;
  budgetReadOnly?: boolean;
  expiresAtReadOnly?: boolean;
  budgetUsage?: number;
  isNewConnection: boolean;
}

const Permissions: React.FC<PermissionsProps> = ({
  capabilities,
  permissions,
  setPermissions,
  isNewConnection,
  budgetUsage,
  readOnly,
  scopesReadOnly,
  budgetReadOnly,
  expiresAtReadOnly,
}) => {
  const [showBudgetOptions, setShowBudgetOptions] = React.useState(
    permissions.scopes.includes("pay_invoice") && permissions.maxAmount > 0
  );
  const [showExpiryOptions, setShowExpiryOptions] = React.useState(
    !!permissions.expiresAt
  );

  const handlePermissionsChange = React.useCallback(
    (changedPermissions: Partial<AppPermissions>) => {
      setPermissions((currentPermissions) => ({
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
    <div className="max-w-lg">
      {permissions.isolated && (
        <p className="mb-4">
          This app is isolated from the rest of your wallet. This means it will
          have an isolated balance and only has access to its own transaction
          history. It will not be able to sign messages on your node's behalf.
        </p>
      )}

      {!readOnly && !scopesReadOnly ? (
        <Scopes
          capabilities={capabilities}
          scopes={permissions.scopes}
          isolated={permissions.isolated}
          onScopesChanged={onScopesChanged}
          isNewConnection={isNewConnection}
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

      {!permissions.isolated && permissions.scopes.includes("pay_invoice") && (
        <>
          {!readOnly && !budgetReadOnly ? (
            <>
              {!showBudgetOptions && (
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => {
                    handleBudgetRenewalChange("monthly");
                    handleBudgetMaxAmountChange(100_000);
                    setShowBudgetOptions(true);
                  }}
                  className={cn("mr-4", showExpiryOptions && "mb-4")}
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
                    onClose={() => {
                      handleBudgetRenewalChange("never");
                      handleBudgetMaxAmountChange(0);
                      setShowBudgetOptions(false);
                    }}
                  />
                  <BudgetAmountSelect
                    value={permissions.maxAmount}
                    onChange={handleBudgetMaxAmountChange}
                  />
                </>
              )}
            </>
          ) : (
            <div className="pl-4 ml-2 border-l-2 border-l-primary mb-4">
              <div className="flex flex-col gap-2 text-muted-foreground text-sm">
                <p className="capitalize">
                  <span className="text-primary font-medium">
                    Budget Renewal:
                  </span>{" "}
                  {permissions.budgetRenewal || "Never"}
                </p>
                <p>
                  <span className="text-primary font-medium">
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
          )}
        </>
      )}

      {!permissions.isolated && (
        <>
          {!readOnly && !expiresAtReadOnly ? (
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
          ) : (
            <>
              <p className="text-sm font-medium mb-2">Connection expiry</p>
              <p className="text-muted-foreground text-sm">
                {permissions.expiresAt
                  ? new Date(permissions.expiresAt).toString()
                  : "This app will never expire"}
              </p>
            </>
          )}
        </>
      )}
      <Card className="mt-5">
        <CardHeader className="flex flex-row justify-between items-center">
          <div>
            <CardTitle>💰️ Support Amethyst with 1000 sats / month</CardTitle>
            <CardDescription>
              Setup a recurring payment to support developers
            </CardDescription>
          </div>
          <Switch />
        </CardHeader>
      </Card>
    </div>
  );
};

export default Permissions;
