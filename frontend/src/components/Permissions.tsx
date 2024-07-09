import { format } from "date-fns";
import { CalendarIcon, PlusCircle, XIcon } from "lucide-react";
import React, { useState } from "react";
import Scopes from "src/components/Scopes";
import { Button } from "src/components/ui/button";
import { Calendar } from "src/components/ui/calendar";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "src/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { cn } from "src/lib/utils";
import {
  AppPermissions,
  BudgetRenewalType,
  NIP_47_PAY_INVOICE_METHOD,
  Scope,
  WalletCapabilities,
  budgetOptions,
  expiryOptions,
  iconMap,
  scopeDescriptions,
  validBudgetRenewals,
} from "src/types";

const getTimeZoneDirection = () => {
  const offset = new Date().getTimezoneOffset();

  return offset <= 0 ? +1 : -1;
};

const daysFromNow = (date?: Date) => {
  if (!date) {
    return 0;
  }
  const utcDate = new Date(
    Date.UTC(
      date.getUTCFullYear(),
      date.getUTCMonth(),
      date.getUTCDate(),
      23,
      59,
      59,
      0
    )
  );
  return Math.ceil(
    (new Date(utcDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24)
  );
};

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
  const [expiryDays, setExpiryDays] = useState(
    daysFromNow(permissions.expiresAt)
  );
  const [budgetOption, setBudgetOption] = useState(
    isNewConnection ? !!permissions.maxAmount : true
  );
  const [customBudget, setCustomBudget] = useState(
    permissions.maxAmount
      ? !Object.values(budgetOptions).includes(permissions.maxAmount)
      : false
  );
  const [expireOption, setExpireOption] = useState(
    isNewConnection ? !!permissions.expiresAt : true
  );
  const [customExpiry, setCustomExpiry] = useState(
    daysFromNow(permissions.expiresAt)
      ? !Object.values(expiryOptions).includes(
          daysFromNow(permissions.expiresAt)
        )
      : false
  );

  // this is triggered when edit mode is cancelled in show app
  React.useEffect(() => {
    setPermissions(initialPermissions);
    setExpiryDays(daysFromNow(initialPermissions.expiresAt));
    setCustomBudget(
      initialPermissions.maxAmount
        ? !Object.values(budgetOptions).includes(initialPermissions.maxAmount)
        : false
    );
    setCustomExpiry(
      daysFromNow(initialPermissions.expiresAt)
        ? !Object.values(expiryOptions).includes(
            daysFromNow(initialPermissions.expiresAt)
          )
        : false
    );
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
    setExpiryDays(expiryDays + getTimeZoneDirection());
    if (!expiryDays) {
      handlePermissionsChange({ expiresAt: undefined });
      return;
    }
    const currentDate = new Date();
    const expiryDate = new Date(
      Date.UTC(
        currentDate.getUTCFullYear(),
        currentDate.getUTCMonth(),
        currentDate.getUTCDate() + expiryDays,
        23,
        59,
        59,
        0
      )
    );
    handlePermissionsChange({ expiresAt: expiryDate });
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
          {expiryDays &&
          permissions.expiresAt &&
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
                  <Select
                    value={permissions.budgetRenewal || "never"}
                    onValueChange={(value) =>
                      handleBudgetRenewalChange(value as BudgetRenewalType)
                    }
                  >
                    <SelectTrigger className="w-[150px] capitalize">
                      <SelectValue placeholder={permissions.budgetRenewal} />
                    </SelectTrigger>
                    <SelectContent className="capitalize">
                      {validBudgetRenewals.map((renewalOption) => (
                        <SelectItem key={renewalOption} value={renewalOption}>
                          {renewalOption}
                        </SelectItem>
                      ))}
                    </SelectContent>
                    <XIcon
                      className="cursor-pointer w-4 text-muted-foreground"
                      onClick={() => handleBudgetRenewalChange("never")}
                    />
                  </Select>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-5 gap-2 text-xs mb-4">
                  {Object.keys(budgetOptions).map((budget) => {
                    return (
                      // replace with something else and then remove dark prefixes
                      <div
                        key={budget}
                        onClick={() => {
                          setCustomBudget(false);
                          handleBudgetMaxAmountChange(budgetOptions[budget]);
                        }}
                        className={cn(
                          "cursor-pointer rounded text-nowrap border-2 text-center p-4 dark:text-white",
                          !customBudget &&
                            (Number.isNaN(permissions.maxAmount)
                              ? 100000
                              : +permissions.maxAmount) == budgetOptions[budget]
                            ? "border-primary"
                            : "border-muted"
                        )}
                      >
                        {`${budget} ${budgetOptions[budget] ? " sats" : ""}`}
                      </div>
                    );
                  })}
                  <div
                    onClick={() => {
                      setCustomBudget(true);
                      handleBudgetMaxAmountChange(0);
                    }}
                    className={cn(
                      "cursor-pointer rounded border-2 text-center p-4 dark:text-white",
                      customBudget ? "border-primary" : "border-muted"
                    )}
                  >
                    Custom...
                  </div>
                </div>
                {customBudget && (
                  <div className="w-full mb-6">
                    <Label htmlFor="budget" className="block mb-2">
                      Custom budget amount (sats)
                    </Label>
                    <Input
                      id="budget"
                      name="budget"
                      type="number"
                      required
                      min={1}
                      value={permissions.maxAmount || ""}
                      onChange={(e) => {
                        handleBudgetMaxAmountChange(parseInt(e.target.value));
                      }}
                    />
                  </div>
                )}
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
        <div className="mb-6">
          <p className="font-medium text-sm mb-2">Connection expiration</p>
          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 text-xs mb-4">
            {Object.keys(expiryOptions).map((expiry) => {
              return (
                <div
                  key={expiry}
                  onClick={() => {
                    setCustomExpiry(false);
                    handleExpiryDaysChange(
                      expiryOptions[expiry] - getTimeZoneDirection()
                    );
                  }}
                  className={cn(
                    "cursor-pointer rounded text-nowrap border-2 text-center p-4 dark:text-white",
                    !customExpiry && expiryDays == expiryOptions[expiry]
                      ? "border-primary"
                      : "border-muted"
                  )}
                >
                  {expiry}
                </div>
              );
            })}
            <Popover>
              <PopoverTrigger asChild>
                <div
                  onClick={() => {}}
                  className={cn(
                    "flex items-center justify-center md:col-span-2 cursor-pointer rounded text-nowrap border-2 text-center px-3 py-4 dark:text-white",
                    customExpiry ? "border-primary" : "border-muted"
                  )}
                >
                  <CalendarIcon className="mr-2 h-4 w-4" />
                  <span className="truncate">
                    {customExpiry && permissions.expiresAt
                      ? format(permissions.expiresAt, "PPP")
                      : "Custom..."}
                  </span>
                </div>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0">
                <Calendar
                  mode="single"
                  disabled={{
                    before: new Date(),
                  }}
                  selected={permissions.expiresAt}
                  onSelect={(date?: Date) => {
                    if (!date) {
                      return;
                    }
                    const dateBefore = new Date(date);
                    dateBefore.setDate(
                      dateBefore.getDate() - getTimeZoneDirection()
                    );
                    setCustomExpiry(true);
                    handleExpiryDaysChange(daysFromNow(dateBefore));
                  }}
                  initialFocus
                />
              </PopoverContent>
            </Popover>
          </div>
        </div>
      )}
    </div>
  );
};

export default Permissions;
