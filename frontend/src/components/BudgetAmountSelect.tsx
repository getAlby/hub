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
  return (
    <>
      <div className="grid grid-cols-3 gap-3 text-xs mb-3">
        {Object.keys(budgetOptions)
          .filter((budget) => !minAmount || budgetOptions[budget] > minAmount)
          .map((budget) => (
            <button
              type="button"
              key={budget}
              onClick={() => {
                onChange(budgetOptions[budget]);
              }}
              className={cn(
                "cursor-pointer rounded text-nowrap border-2 text-center p-3 py-4 slashed-zero",
                value == budgetOptions[budget]
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
      <div className="mb-4">
        <Input
          id="budget"
          name="budget"
          type="number"
          required
          placeholder="Custom amount in sats"
          min={minAmount || 1}
          value={value || ""}
          onChange={(e) => {
            onChange(parseInt(e.target.value));
          }}
        />
      </div>
    </>
  );
}

export default BudgetAmountSelect;
