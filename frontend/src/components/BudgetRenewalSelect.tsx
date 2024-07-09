import React from "react";
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
  disabled?: boolean;
}

const BudgetRenewalSelect: React.FC<BudgetRenewalProps> = ({
  value,
  onChange,
  disabled,
}) => {
  return (
    <Select value={value} onValueChange={onChange} disabled={disabled}>
      <SelectTrigger className="w-[150px]">
        <SelectValue placeholder={"placeholder"} />
      </SelectTrigger>
      <SelectContent>
        {validBudgetRenewals.map((renewalOption) => (
          <SelectItem key={renewalOption} value={renewalOption}>
            {renewalOption.charAt(0).toUpperCase() + renewalOption.slice(1)}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
};

export default BudgetRenewalSelect;
