import React from "react";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { cn } from "src/lib/utils";
import { budgetOptions } from "src/types";

function BudgetAmountSelect({
  value,
  onChange,
}: {
  value: number;
  onChange: (value: number) => void;
}) {
  const [customBudget, setCustomBudget] = React.useState(
    value ? !Object.values(budgetOptions).includes(value) : false
  );
  return (
    <>
      <div className="grid grid-cols-2 md:grid-cols-5 gap-2 text-xs mb-4">
        {Object.keys(budgetOptions).map((budget) => {
          return (
            // replace with something else and then remove dark prefixes
            <div
              key={budget}
              onClick={() => {
                setCustomBudget(false);
                onChange(budgetOptions[budget]);
              }}
              className={cn(
                "cursor-pointer rounded text-nowrap border-2 text-center p-4 dark:text-white",
                !customBudget &&
                  (Number.isNaN(value) ? 100000 : value) ==
                    budgetOptions[budget]
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
            onChange(0);
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
            value={value || ""}
            onChange={(e) => {
              onChange(parseInt(e.target.value));
            }}
          />
        </div>
      )}
      {/* <table className="text-muted-foreground">
        <tbody>
          <tr className="text-sm">
            <td className="pr-2">Budget Allowance:</td>
            <td>
              {value ? new Intl.NumberFormat().format(value) : "∞"}
              {" sats "}
              ({new Intl.NumberFormat().format(budgetUsage || 0)} sats
              used)
            </td>
          </tr>
          <tr className="text-sm">
            <td className="pr-2">Renews:</td>
            <td className="capitalize">
              {permissions.budgetRenewal || "Never"}
            </td>
          </tr>
        </tbody>
      </table> */}
    </>
  );
}

export default BudgetAmountSelect;
