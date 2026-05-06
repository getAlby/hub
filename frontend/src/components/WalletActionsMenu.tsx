import {
  ArrowDownUpIcon,
  CalendarSyncIcon,
  CreditCardIcon,
  DownloadIcon,
  EllipsisVerticalIcon,
} from "lucide-react";
import { useNavigate } from "react-router";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { ProDropdownMenuItem } from "src/components/UpgradeDialog";
import { openLink } from "src/utils/openLink";
import { handleExportTransactions } from "./transactions-utils";

export function WalletActionsMenu({
  hasChannelManagement,
}: {
  hasChannelManagement: boolean;
}) {
  const navigate = useNavigate();

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
            <DropdownMenuItem onClick={() => navigate("/wallet/swap")}>
              <ArrowDownUpIcon className="h-4 w-4" />
              Swap
            </DropdownMenuItem>
          )}
          <DropdownMenuItem
            onClick={() => navigate("/internal-apps/zapplanner")}
          >
            <CalendarSyncIcon className="h-4 w-4" />
            Recurring
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => openLink("https://www.getalby.com/topup")}
          >
            <CreditCardIcon className="h-4 w-4" />
            Buy
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
