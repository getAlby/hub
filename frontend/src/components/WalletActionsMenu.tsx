import {
  ArrowDownUpIcon,
  CalendarSyncIcon,
  CreditCardIcon,
  DownloadIcon,
  EllipsisVerticalIcon,
  FunnelIcon,
} from "lucide-react";
import { useState } from "react";
import { Link } from "react-router";
import ExternalLink from "src/components/ExternalLink";
import { TransactionsFilterDialog } from "src/components/TransactionsFilterDialog";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { ProDropdownMenuItem } from "src/components/UpgradeDialog";
import useTransactionFiltersStore from "src/state/TransactionFiltersStore";
import { handleExportTransactions } from "./transactions-utils";

export function WalletActionsMenu({
  hasChannelManagement,
  isOnchain,
}: {
  hasChannelManagement: boolean;
  isOnchain: boolean;
}) {
  const [filterDialogOpen, setFilterDialogOpen] = useState(false);
  const { filters, setFilters } = useTransactionFiltersStore();

  return (
    <>
      <DropdownMenu>
        <Button
          asChild
          size="icon"
          variant="ghost"
          className={isOnchain ? "sm:hidden" : undefined}
        >
          <DropdownMenuTrigger>
            <EllipsisVerticalIcon className="h-4 w-4" />
          </DropdownMenuTrigger>
        </Button>
        <DropdownMenuContent align="end">
          <div className="sm:hidden">
            {hasChannelManagement && (
              <DropdownMenuItem asChild>
                <Link to="/wallet/swap" className="w-full cursor-pointer">
                  <ArrowDownUpIcon className="h-4 w-4" />
                  Swap
                </Link>
              </DropdownMenuItem>
            )}
            <DropdownMenuItem asChild>
              <Link
                to="/internal-apps/zapplanner"
                className="w-full cursor-pointer"
              >
                <CalendarSyncIcon className="h-4 w-4" />
                Recurring
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <ExternalLink
                to="https://www.getalby.com/topup"
                className="w-full cursor-pointer"
              >
                <CreditCardIcon className="h-4 w-4" />
                Buy
              </ExternalLink>
            </DropdownMenuItem>
            {!isOnchain && <DropdownMenuSeparator />}
          </div>
          {!isOnchain && (
            <>
              <DropdownMenuItem onClick={() => setFilterDialogOpen(true)}>
                <FunnelIcon className="h-4 w-4" />
                Filter Transactions
              </DropdownMenuItem>
              <ProDropdownMenuItem onClick={() => handleExportTransactions()}>
                <DownloadIcon className="h-4 w-4" />
                Export Transactions
              </ProDropdownMenuItem>
            </>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
      {!isOnchain && (
        <TransactionsFilterDialog
          open={filterDialogOpen}
          onOpenChange={setFilterDialogOpen}
          filters={filters}
          onFiltersChange={setFilters}
        />
      )}
    </>
  );
}
