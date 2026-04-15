import { DownloadIcon, EllipsisVerticalIcon } from "lucide-react";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { ProDropdownMenuItem } from "src/components/UpgradeDialog";
import { handleExportTransactions } from "./transactions-utils";

export const TransactionsListMenu = ({ appId }: { appId?: number }) => {
  return (
    <DropdownMenu>
      <Button asChild size="icon" variant="ghost">
        <DropdownMenuTrigger>
          <EllipsisVerticalIcon className="h-4 w-4" />
        </DropdownMenuTrigger>
      </Button>
      <DropdownMenuContent align="end">
        <ProDropdownMenuItem onClick={() => handleExportTransactions(appId)}>
          <DownloadIcon className="h-4 w-4" />
          Export Transactions
        </ProDropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
