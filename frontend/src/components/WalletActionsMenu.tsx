import {
  ArrowDownUpIcon,
  CalendarSyncIcon,
  CreditCardIcon,
  DownloadIcon,
  EllipsisVerticalIcon,
} from "lucide-react";
import { Link } from "react-router";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { ProDropdownMenuItem } from "src/components/UpgradeDialog";
import { handleExportTransactions } from "./transactions-utils";

export function WalletActionsMenu({
  hasChannelManagement,
}: {
  hasChannelManagement: boolean;
}) {
  return (
    <DropdownMenu>
      <Button asChild size="icon" variant="ghost">
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
          <DropdownMenuSeparator />
        </div>
        <ProDropdownMenuItem onClick={() => handleExportTransactions()}>
          <DownloadIcon className="h-4 w-4" />
          Export Transactions
        </ProDropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
