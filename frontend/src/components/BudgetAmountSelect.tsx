import React from "react";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
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
              <button
                type="button"
                key={budget}
                onClick={() => {
                  setCustomBudget(false);
                  onChange(budgetOptions[budget]);
                }}
                className={cn(
                  "cursor-pointer rounded text-nowrap border-2 text-center p-2 py-4 slashed-zero",
                  !customBudget && value == budgetOptions[budget]
                    ? "border-primary"
                    : "border-muted"
                )}
              >
                {`${budget} ${budgetOptions[budget] ? " sats" : ""}`}

                {budgetOptions[budget] > 0 && (
                  <FormattedFiatAmount
                    amount={budgetOptions[budget]}
                    className="text-xs"
                    showApprox
                  />
                )}
              </button>
            );
          })}
        <button
          onClick={() => {
            setCustomBudget(true);
            onChange(0);
          }}
          className={cn(
            "cursor-pointer rounded border-2 text-center p-4 dark:text-white",
            customBudget ? "border-primary" : "border-muted"
          )}
        >
          Custom
        </button>
      </div>
      {customBudget && (
        <div className="grid gap-2 mb-5">
          <Label htmlFor="budget">Custom budget amount (sats)</Label>
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
