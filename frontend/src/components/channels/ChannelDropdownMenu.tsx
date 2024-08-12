import {
  ExternalLinkIcon,
  HandCoins,
  MoreHorizontal,
  Trash2,
} from "lucide-react";
import React from "react";
import { CloseChannelDialogContent } from "src/components/CloseChannelDialogContent";
import ExternalLink from "src/components/ExternalLink";
import { RoutingFeeDialogContent } from "src/components/RoutingFeeDialogContent";
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
};

export function ChannelDropdownMenu({
  alias,
  channel,
}: ChannelDropdownMenuProps) {
  const [dialog, setDialog] = React.useState<"closeChannel" | "routingFee">();

  const showCloseChannelDialog = () => setDialog("closeChannel");
  const showRoutingFeeDialog = () => setDialog("routingFee");

  return (
    <AlertDialog
      onOpenChange={() => {
        if (!open) {
          setDialog(undefined);
        }
      }}
    >
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
            <AlertDialogTrigger asChild>
              <DropdownMenuItem
                className="flex flex-row items-center gap-2 cursor-pointer"
                onClick={showRoutingFeeDialog}
              >
                <HandCoins className="h-4 w-4" />
                Set Routing Fee
              </DropdownMenuItem>
            </AlertDialogTrigger>
          )}
          <AlertDialogTrigger asChild>
            <DropdownMenuItem
              className="flex flex-row items-center gap-2 cursor-pointer"
              onClick={showCloseChannelDialog}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
              Close Channel
            </DropdownMenuItem>
          </AlertDialogTrigger>
        </DropdownMenuContent>
      </DropdownMenu>
      {dialog === "closeChannel" && (
        <CloseChannelDialogContent alias={alias} channel={channel} />
      )}
      {dialog === "routingFee" && <RoutingFeeDialogContent channel={channel} />}
    </AlertDialog>
  );
}
