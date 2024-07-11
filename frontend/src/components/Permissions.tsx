import { PlusCircle } from "lucide-react";
import React, { useState } from "react";
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
  canEditPermissions: boolean;
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
  const [budgetOption, setBudgetOption] = useState(
    isNewConnection ? !!permissions.maxAmount : true
  );
  const [expireOption, setExpireOption] = useState(
    isNewConnection ? !!permissions.expiresAt : true
  );
  // this is triggered when edit mode is cancelled in show app
  React.useEffect(() => {
    setPermissions(initialPermissions);
  }, [initialPermissions]);

  const handlePermissionsChange = (
    changedPermissions: Partial<AppPermissions>
  ) => {
    const updatedPermissions = { ...permissions, ...changedPermissions };
    setPermissions(updatedPermissions);
    onPermissionsChange(updatedPermissions);
  };

  const handleScopeChange = (scopes: Set<Scope>) => {
    handlePermissionsChange({ scopes });
  };

  const handleBudgetMaxAmountChange = (amount: number) => {
    handlePermissionsChange({ maxAmount: amount });
  };

  const handleBudgetRenewalChange = (value: string) => {
    handlePermissionsChange({ budgetRenewal: value as BudgetRenewalType });
  };

  const handleExpiryDaysChange = (expiryDays: number) => {
    if (!expiryDays) {
      handlePermissionsChange({ expiresAt: undefined });
      return;
    }
    const currentDate = new Date();
    currentDate.setDate(currentDate.getUTCDate() + expiryDays);
    currentDate.setHours(23, 59, 59);
    handlePermissionsChange({ expiresAt: currentDate });
  };

  return !canEditPermissions ? (
    <>
      <p className="text-sm font-medium mb-2">Scopes</p>
      <div className="flex flex-col gap-1">
        {[...initialPermissions.scopes].map((rm, index) => {
          const PermissionIcon = iconMap[rm];
          return (
            <div
              key={index}
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
      {permissions.scopes.has(NIP_47_PAY_INVOICE_METHOD) && (
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
              {`(${new Intl.NumberFormat().format(budgetUsage || 0)} sats used)`}
            </p>
          </div>
        </div>
      )}
      <div className="mt-4">
        <p className="text-sm font-medium mb-2">Connection expiry</p>
        <p className="text-muted-foreground text-sm">
          {permissions.expiresAt &&
          new Date(permissions.expiresAt).getFullYear() !== 1
            ? new Date(permissions.expiresAt).toString()
            : "This app will never expire"}
        </p>
      </div>
    </>
  ) : (
    <div className="max-w-lg">
      <Scopes
        capabilities={capabilities}
        scopes={permissions.scopes}
        onScopeChange={handleScopeChange}
      />

      {capabilities.scopes.includes(NIP_47_PAY_INVOICE_METHOD) &&
        permissions.scopes.has(NIP_47_PAY_INVOICE_METHOD) && (
          <>
            {!budgetOption && (
              <Button
                type="button"
                variant="secondary"
                onClick={() => setBudgetOption(true)}
                className="mb-4 mr-4"
              >
                <PlusCircle className="w-4 h-4 mr-2" />
                Set budget renewal
              </Button>
            )}
            {budgetOption && (
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
        )}

      {!expireOption && (
        <Button
          type="button"
          variant="secondary"
          onClick={() => setExpireOption(true)}
          className="mb-6"
        >
          <PlusCircle className="w-4 h-4 mr-2" />
          Set expiration time
        </Button>
      )}

      {expireOption && (
        <ExpirySelect
          value={permissions.expiresAt}
          onChange={handleExpiryDaysChange}
        />
      )}
    </div>
  );
};

export default Permissions;
