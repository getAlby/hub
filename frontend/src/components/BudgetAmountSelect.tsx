import { budgetOptions } from "src/types";

function BudgetAmountSelect({
  value,
  onChange,
}: {
  value?: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="grid grid-cols-6 grid-rows-2 md:grid-rows-1 md:grid-cols-6 gap-2 text-xs">
      {Object.keys(budgetOptions).map((budget) => {
        const amount = budgetOptions[budget];
        return (
          <div
            key={budget}
            onClick={() => onChange(amount)}
            className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
              value === amount ? "border-primary" : "border-muted"
            } text-center py-4`}
          >
            {budget}
            <br />
            {amount ? "sats" : "#reckless"}
          </div>
        );
      })}
    </div>
  );
}

export default BudgetAmountSelect;
