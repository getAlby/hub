import { XIcon } from "lucide-react";
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
}

const BudgetRenewalSelect: React.FC<BudgetRenewalProps> = ({
  value,
  onChange,
}) => {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger className="w-[150px] capitalize">
        <SelectValue placeholder={value} />
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
        onClick={() => onChange("never")}
      />
    </Select>
  );
};

export default BudgetRenewalSelect;
