import React from "react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Input } from "src/components/ui/input";
import { cn } from "src/lib/utils";
import { budgetOptionsSat as defaultBudgetOptionsSat } from "src/types";

function BudgetAmountSelect({
  valueSat,
  onChange,
  minAmountSat,
  budgetOptionsSat = defaultBudgetOptionsSat,
}: {
  valueSat: number;
  onChange: (value: number) => void;
  minAmountSat?: number;
  budgetOptionsSat?: typeof defaultBudgetOptionsSat;
}) {
  const [inputValue, setInputValue] = React.useState(
    valueSat ? String(valueSat) : ""
  );

  React.useEffect(() => {
    setInputValue(valueSat ? String(valueSat) : "");
  }, [valueSat]);

  return (
    <>
      <div className="grid grid-cols-3 gap-3 text-xs mb-3">
        {Object.keys(budgetOptionsSat)
          .filter(
            (budget) =>
              !minAmountSat || budgetOptionsSat[budget] >= minAmountSat
          )
          .map((budget) => (
            <button
              type="button"
              key={budget}
              onClick={() => {
                onChange(budgetOptionsSat[budget]);
              }}
              className={cn(
                "cursor-pointer rounded text-nowrap border-2 text-center p-3 py-4 slashed-zero",
                valueSat === budgetOptionsSat[budget]
                  ? "border-primary"
                  : "border-muted"
              )}
            >
              <FormattedBitcoinAmount
                amountMsat={budgetOptionsSat[budget] * 1000}
              />
              <FormattedFiatAmount
                className="text-xs"
                showApprox
                amountSat={budgetOptionsSat[budget]}
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
