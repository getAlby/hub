import {
  ExternalLinkIcon,
  HandCoinsIcon,
  MoreHorizontalIcon,
  ScaleIcon,
  Trash2Icon,
} from "lucide-react";
import React from "react";
import { useSearchParams } from "react-router-dom";
import { CloseChannelDialogContent } from "src/components/CloseChannelDialogContent";
import ExternalLink from "src/components/ExternalLink";
import { RebalanceChannelDialogContent } from "src/components/RebalanceChannelDialogContent";
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
import { useInfo } from "src/hooks/useInfo";
import { Channel } from "src/types";

type ChannelDropdownMenuProps = {
  alias: string;
  channel: Channel;
};

export function ChannelDropdownMenu({
  alias,
  channel,
}: ChannelDropdownMenuProps) {
  const { data: info } = useInfo();
  const [searchParams] = useSearchParams();
  const [dialog, setDialog] = React.useState<
    "closeChannel" | "routingFee" | "rebalance"
  >();

  React.useEffect(() => {
    // when opening the swap dialog, close existing dialog
    if (searchParams.has("swap", "true")) {
      setDialog(undefined);
    }
  }, [searchParams]);

  return (
    <AlertDialog
      onOpenChange={(open) => {
        if (!open) {
          setDialog(undefined);
        }
      }}
    >
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button size="icon" variant="ghost">
            <MoreHorizontalIcon className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {channel.status == "online" && (
            <AlertDialogTrigger asChild>
              <DropdownMenuItem
                className="flex flex-row items-center gap-2 cursor-pointer"
                onClick={() => setDialog("rebalance")}
              >
                <ScaleIcon className="h-4 w-4" />
                Rebalance In
              </DropdownMenuItem>
            </AlertDialogTrigger>
          )}
          <DropdownMenuItem className="flex flex-row items-center gap-2 cursor-pointer">
            <ExternalLink
              to={`${info?.mempoolUrl}/tx/${channel.fundingTxId}#flow=&vout=${channel.fundingTxVout}`}
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
                <HandCoinsIcon className="h-4 w-4" />
                Set Routing Fee
              </DropdownMenuItem>
            </AlertDialogTrigger>
          )}
          <AlertDialogTrigger asChild>
            <DropdownMenuItem
              className="flex flex-row items-center gap-2 cursor-pointer"
              onClick={() => setDialog("closeChannel")}
            >
              <Trash2Icon className="h-4 w-4 text-destructive" />
              Close Channel
            </DropdownMenuItem>
          </AlertDialogTrigger>
        </DropdownMenuContent>
      </DropdownMenu>
      {dialog === "closeChannel" && (
        <CloseChannelDialogContent alias={alias} channel={channel} />
      )}
      {dialog === "routingFee" && <RoutingFeeDialogContent channel={channel} />}
      {dialog === "rebalance" && (
        <RebalanceChannelDialogContent
          receiveThroughNodePubkey={channel.remotePubkey}
          closeDialog={() => setDialog(undefined)}
        />
      )}
    </AlertDialog>
  );
}
