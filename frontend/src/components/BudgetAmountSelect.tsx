import React from "react";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { cn } from "src/lib/utils";
import { budgetOptions as defaultBudgetOptions } from "src/types";

function BudgetAmountSelect({
  value,
  onChange,
  minAmount,
  budgetOptions = defaultBudgetOptions,
}: {
  value: number;
  onChange: (value: number) => void;
  minAmount?: number;
  budgetOptions?: typeof defaultBudgetOptions;
}) {
  const [customBudget, setCustomBudget] = React.useState(
    value ? !Object.values(budgetOptions).includes(value) : false
  );
  return (
    <>
      <div className="grid grid-cols-2 md:grid-cols-5 gap-2 text-xs mb-4">
        {Object.keys(budgetOptions)
          .filter(
            (budget) =>
              !minAmount ||
              !budgetOptions[budget] ||
              budgetOptions[budget] > minAmount
          )
          .map((budget) => {
            return (
              <div
                key={budget}
                onClick={() => {
                  setCustomBudget(false);
                  onChange(budgetOptions[budget]);
                }}
                className={cn(
                  "cursor-pointer rounded text-nowrap border-2 text-center p-4 slashed-zero",
                  !customBudget && value == budgetOptions[budget]
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
            autoFocus
            min={minAmount || 1}
            value={value || ""}
            onChange={(e) => {
              onChange(parseInt(e.target.value));
            }}
          />
        </div>
      )}
    </>
  );
}

export default BudgetAmountSelect;
