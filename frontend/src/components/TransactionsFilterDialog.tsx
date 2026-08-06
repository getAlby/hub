import React from "react";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { ToggleGroup, ToggleGroupItem } from "src/components/ui/toggle-group";
import {
  defaultTransactionFilters,
  type TransactionFilters,
} from "src/hooks/useTransactions";

const TYPE_OPTIONS: { label: string; value: string }[] = [
  { label: "All", value: "all" },
  { label: "Sent", value: "outgoing" },
  { label: "Received", value: "incoming" },
];

type TransactionsFilterDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  filters: TransactionFilters;
  onFiltersChange: (filters: TransactionFilters) => void;
};

export function TransactionsFilterDialog({
  open,
  onOpenChange,
  filters,
  onFiltersChange,
}: TransactionsFilterDialogProps) {
  const [searchTerm, setSearchTerm] = React.useState("");
  const [type, setType] = React.useState("all");
  const [minAmountSat, setMinAmountSat] = React.useState("");
  const [hideFailed, setHideFailed] = React.useState(false);

  React.useEffect(() => {
    if (open) {
      setSearchTerm(filters.searchTerm ?? "");
      setType(filters.type ?? "all");
      setMinAmountSat(filters.minAmountSat ? String(filters.minAmountSat) : "");
      setHideFailed(!!filters.hideFailed);
    }
  }, [open, filters]);

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const parsedMinAmountSat = Number(minAmountSat);
    onFiltersChange({
      searchTerm: searchTerm.trim() || undefined,
      type: type === "incoming" || type === "outgoing" ? type : undefined,
      minAmountSat:
        Number.isSafeInteger(parsedMinAmountSat) && parsedMinAmountSat > 0
          ? parsedMinAmountSat
          : undefined,
      hideFailed,
    });
    onOpenChange(false);
  }

  function onReset() {
    onFiltersChange({ ...defaultTransactionFilters });
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={onSubmit}>
          <DialogHeader>
            <DialogTitle>Filter Transactions</DialogTitle>
            <DialogDescription>
              Choose which payments appear in your transaction list.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-2 mt-5">
            <Label htmlFor="searchTerm">Search</Label>
            <Input
              autoFocus
              id="searchTerm"
              type="text"
              placeholder="Description, payment hash, invoice or label"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div className="grid gap-2 mt-4">
            <Label>Direction</Label>
            <ToggleGroup
              type="single"
              variant="outline"
              value={type}
              onValueChange={(value) => value && setType(value)}
            >
              {TYPE_OPTIONS.map((option) => (
                <ToggleGroupItem key={option.value} value={option.value}>
                  {option.label}
                </ToggleGroupItem>
              ))}
            </ToggleGroup>
          </div>
          <div className="grid gap-2 mt-4">
            <Label htmlFor="minAmountSat">Minimum amount (sats)</Label>
            <Input
              id="minAmountSat"
              type="number"
              min="1"
              placeholder="Show all amounts"
              value={minAmountSat}
              onChange={(e) => setMinAmountSat(e.target.value.trim())}
            />
          </div>
          <div className="flex items-center mt-4">
            <Checkbox
              id="hideFailed"
              checked={hideFailed}
              onCheckedChange={(checked) => setHideFailed(checked === true)}
            />
            <Label htmlFor="hideFailed" className="ml-2 cursor-pointer">
              Hide failed payments
            </Label>
          </div>
          <DialogFooter className="mt-5">
            <Button type="button" variant="secondary" onClick={onReset}>
              Reset
            </Button>
            <Button type="submit">Apply Filters</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
