import React from "react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Input } from "src/components/ui/input";
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
  const [inputValue, setInputValue] = React.useState(
    value ? String(value) : ""
  );

  React.useEffect(() => {
    setInputValue(value ? String(value) : "");
  }, [value]);

  return (
    <>
      <div className="grid grid-cols-3 gap-3 text-xs mb-3">
        {Object.keys(budgetOptions)
          .filter((budget) => !minAmount || budgetOptions[budget] >= minAmount)
          .map((budget) => (
            <button
              type="button"
              key={budget}
              onClick={() => {
                onChange(budgetOptions[budget]);
              }}
              className={cn(
                "cursor-pointer rounded text-nowrap border-2 text-center p-3 py-4 slashed-zero",
                value === budgetOptions[budget]
                  ? "border-primary"
                  : "border-muted"
              )}
            >
              <FormattedBitcoinAmount amount={budgetOptions[budget] * 1000} />
              <FormattedFiatAmount
                className="text-xs"
                showApprox
                amount={budgetOptions[budget]}
              />
            </button>
          ))}
      </div>
      <div className="mb-3">
        <Input
          id="budget"
          name="budget"
          type="number"
          min={1}
          required
          placeholder="Custom amount in sats"
          value={inputValue}
          onChange={(e) => {
            setInputValue(e.target.value);
            const n = parseInt(e.target.value);
            onChange(!isNaN(n) && n > 0 ? n : 0);
          }}
        />
      </div>
    </>
  );
}

export default BudgetAmountSelect;
