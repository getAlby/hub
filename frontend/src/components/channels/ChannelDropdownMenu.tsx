import {
  ArrowRightLeft,
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
import { openLink } from "src/utils/openLink";

type ChannelDropdownMenuProps = {
  alias: string;
  channel: Channel;
};

export function ChannelDropdownMenu({
  alias,
  channel,
}: ChannelDropdownMenuProps) {
  const [dialog, setDialog] = React.useState<"closeChannel" | "routingFee">();

  return (
    <AlertDialog
      onOpenChange={() => {
        if (!open) {
          setDialog(undefined);
        }
      }}
    >
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button size="icon" variant="ghost">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem className="flex flex-row items-center gap-2 cursor-pointer">
            <ExternalLink
              to={`https://mempool.space/tx/${channel.fundingTxId}#flow=&vout=${channel.fundingTxVout}`}
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
                onClick={() => setDialog("routingFee")}
              >
                <HandCoins className="h-4 w-4" />
                Set Routing Fee
              </DropdownMenuItem>
            </AlertDialogTrigger>
          )}
          {channel.localSpendableBalance > 0 && (
            <DropdownMenuItem
              className="flex flex-row items-center gap-2 cursor-pointer"
              onClick={() => {
                const amount = channel.localSpendableBalance / 1000;
                // TODO: Fetch new onchain address
                const destination =
                  "bc1p2wsldez5mud2yam29q22wgfh9439spgduvct83k3pm50fcxa5dps59h4z5";
                openLink(
                  `https://boltz.exchange/?sendAsset=LN&receiveAsset=BTC&sendAmount=${amount}&destination=${destination}`
                );
              }}
            >
              <ArrowRightLeft className="h-4 w-4" />
              Swap out
            </DropdownMenuItem>
          )}
          <AlertDialogTrigger asChild>
            <DropdownMenuItem
              className="flex flex-row items-center gap-2 cursor-pointer"
              onClick={() => setDialog("closeChannel")}
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
