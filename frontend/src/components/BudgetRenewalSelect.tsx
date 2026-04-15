import React from "react";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { BudgetRenewalType, validBudgetRenewals } from "src/types";

interface BudgetRenewalProps {
  value: BudgetRenewalType;
  onChange: (value: BudgetRenewalType) => void;
}

const BudgetRenewalSelect: React.FC<BudgetRenewalProps> = ({
  value,
  onChange,
}) => {
  return (
    <div className="flex gap-3 items-center mb-3">
      <Label htmlFor="budget-renewal">Renewal</Label>
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger id="budget-renewal" className="w-[150px] capitalize">
          <SelectValue placeholder={value} />
        </SelectTrigger>
        <SelectContent className="capitalize">
          {validBudgetRenewals.map((renewalOption) => (
            <SelectItem key={renewalOption} value={renewalOption}>
              {renewalOption}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
};

export default BudgetRenewalSelect;
