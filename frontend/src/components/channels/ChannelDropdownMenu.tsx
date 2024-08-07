import {
  ExternalLinkIcon,
  HandCoins,
  MoreHorizontal,
  Trash2,
} from "lucide-react";
import { CloseChannelDialog } from "src/components/CloseChannelDialog";
import ExternalLink from "src/components/ExternalLink";
import {
  AlertDialog,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button.tsx";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu.tsx";
import { Channel } from "src/types";

type ChannelDropdownMenuProps = {
  alias: string;
  channel: Channel;
  editChannel(channel: Channel): void;
};

export function ChannelDropdownMenu({
  alias,
  channel,
  editChannel,
}: ChannelDropdownMenuProps) {
  return (
    <AlertDialog>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button size="icon" variant="ghost">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem className="flex flex-row items-center gap-2 cursor-pointer">
            <ExternalLink
              to={`https://mempool.space/tx/${channel.fundingTxId}`}
              className="w-full flex flex-row items-center gap-2"
            >
              <ExternalLinkIcon className="w-4 h-4" />
              <p>View Funding Transaction</p>
            </ExternalLink>
          </DropdownMenuItem>
          <DropdownMenuItem className="flex flex-row items-center gap-2 cursor-pointer">
            <ExternalLink
              to={`https://amboss.space/node/${channel.remotePubkey}`}
              className="w-full flex flex-row items-center gap-2"
            >
              <ExternalLinkIcon className="w-4 h-4" />
              <p>View Node on amboss.space</p>
            </ExternalLink>
          </DropdownMenuItem>
          {channel.public && (
            <DropdownMenuItem
              className="flex flex-row items-center gap-2 cursor-pointer"
              onClick={() => editChannel(channel)}
            >
              <HandCoins className="h-4 w-4" />
              Set Routing Fee
            </DropdownMenuItem>
          )}
          <AlertDialogTrigger asChild>
            <DropdownMenuItem className="flex flex-row items-center gap-2 cursor-pointer">
              <Trash2 className="h-4 w-4 text-destructive" />
              Close Channel
            </DropdownMenuItem>
          </AlertDialogTrigger>
        </DropdownMenuContent>
      </DropdownMenu>
      <CloseChannelDialog alias={alias} channel={channel} />
    </AlertDialog>
  );
}
