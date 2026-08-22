import { DownloadIcon, EllipsisVerticalIcon, FunnelIcon } from "lucide-react";
import { useState } from "react";
import { TransactionsFilterDialog } from "src/components/TransactionsFilterDialog";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { ProDropdownMenuItem } from "src/components/UpgradeDialog";
import useTransactionFiltersStore from "src/state/TransactionFiltersStore";
import { handleExportTransactions } from "./transactions-utils";

export const TransactionsListMenu = ({ appId }: { appId?: number }) => {
  const [filterDialogOpen, setFilterDialogOpen] = useState(false);
  const { filters, setFilters } = useTransactionFiltersStore();

  return (
    <>
      <DropdownMenu>
        <Button asChild size="icon" variant="ghost">
          <DropdownMenuTrigger>
            <EllipsisVerticalIcon className="h-4 w-4" />
          </DropdownMenuTrigger>
        </Button>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => setFilterDialogOpen(true)}>
            <FunnelIcon className="h-4 w-4" />
            Filter Transactions
          </DropdownMenuItem>
          <ProDropdownMenuItem onClick={() => handleExportTransactions(appId)}>
            <DownloadIcon className="h-4 w-4" />
            Export Transactions
          </ProDropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <TransactionsFilterDialog
        open={filterDialogOpen}
        onOpenChange={setFilterDialogOpen}
        filters={filters}
        onFiltersChange={setFilters}
      />
    </>
  );
};
