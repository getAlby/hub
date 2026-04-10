import { CheckIcon, FilterIcon } from "lucide-react";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";

const FILTER_OPTIONS = [
  { label: "All Transactions", value: undefined },
  { label: "10 sats", value: 10 },
  { label: "100 sats", value: 100 },
  { label: "1,000 sats", value: 1000 },
  { label: "10,000 sats", value: 10000 },
  { label: "100,000 sats", value: 100000 },
];

type TransactionsFilterMenuProps = {
  selectedFilter: number | undefined;
  onFilterChange: (value: number | undefined) => void;
};

export const TransactionsFilterMenu = ({
  selectedFilter,
  onFilterChange,
}: TransactionsFilterMenuProps) => {
  return (
    <DropdownMenu>
      <Button asChild size="icon" variant="secondary">
        <DropdownMenuTrigger>
          <FilterIcon className="h-4 w-4" />
        </DropdownMenuTrigger>
      </Button>
      <DropdownMenuContent align="end">
        {FILTER_OPTIONS.map((option) => (
          <DropdownMenuItem
            key={option.label}
            className="flex flex-row items-center gap-2 cursor-pointer"
            onClick={() => onFilterChange(option.value)}
          >
            <div className="w-4 h-4 flex items-center justify-center">
              {selectedFilter === option.value && (
                <CheckIcon className="h-4 w-4" />
              )}
            </div>
            <span>{option.label}</span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
