import { DownloadIcon, EllipsisVerticalIcon, FilterXIcon } from "lucide-react";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { ProDropdownMenuItem } from "src/components/UpgradeDialog";
import {
  defaultTransactionFilters,
  type TransactionFilters,
} from "src/hooks/useTransactions";
import { handleExportTransactions } from "./transactions-utils";

const MIN_AMOUNT_FILTER_OPTIONS = [
  { label: "All amounts", value: "all" },
  { label: "10 sats and above", value: "10" },
  { label: "100 sats and above", value: "100" },
  { label: "1,000 sats and above", value: "1000" },
  { label: "10,000 sats and above", value: "10000" },
  { label: "100,000 sats and above", value: "100000" },
];

type TransactionsListMenuProps = {
  appId?: number;
  filters?: TransactionFilters;
  onFiltersChange?: (filters: TransactionFilters) => void;
};

export const TransactionsListMenu = ({
  appId,
  filters,
  onFiltersChange,
}: TransactionsListMenuProps) => {
  const currentFilters = {
    ...defaultTransactionFilters,
    ...filters,
  };
  const hasFilterControls = !!onFiltersChange;
  const hasActiveFilters =
    (currentFilters.minAmountSat ?? 0) > 0 ||
    (currentFilters.showFailed ?? true) === false;

  const setMinAmountSat = (value: string) => {
    onFiltersChange?.({
      ...currentFilters,
      minAmountSat: value === "all" ? undefined : Number(value),
    });
  };

  const setShowFailed = (checked: boolean) => {
    onFiltersChange?.({
      ...currentFilters,
      showFailed: checked,
    });
  };

  return (
    <DropdownMenu>
      <Button
        asChild
        size="icon"
        variant={hasActiveFilters ? "secondary" : "ghost"}
      >
        <DropdownMenuTrigger>
          <EllipsisVerticalIcon className="h-4 w-4" />
        </DropdownMenuTrigger>
      </Button>
      <DropdownMenuContent align="end">
        {hasFilterControls && (
          <>
            <DropdownMenuLabel>Minimum amount</DropdownMenuLabel>
            <DropdownMenuRadioGroup
              value={String(currentFilters.minAmountSat ?? "all")}
              onValueChange={setMinAmountSat}
            >
              {MIN_AMOUNT_FILTER_OPTIONS.map((option) => (
                <DropdownMenuRadioItem key={option.value} value={option.value}>
                  {option.label}
                </DropdownMenuRadioItem>
              ))}
            </DropdownMenuRadioGroup>
            <DropdownMenuSeparator />
            <DropdownMenuCheckboxItem
              checked={currentFilters.showFailed ?? true}
              onCheckedChange={(checked) => setShowFailed(checked === true)}
            >
              Show failed payments
            </DropdownMenuCheckboxItem>
            {hasActiveFilters && (
              <DropdownMenuItem
                onClick={() =>
                  onFiltersChange({ ...defaultTransactionFilters })
                }
              >
                <FilterXIcon className="h-4 w-4" />
                Clear filters
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
          </>
        )}
        <ProDropdownMenuItem onClick={() => handleExportTransactions(appId)}>
          <DownloadIcon className="h-4 w-4" />
          Export Transactions
        </ProDropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
