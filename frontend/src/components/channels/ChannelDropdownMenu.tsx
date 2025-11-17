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
  hasMultipleChannels: boolean;
};

export function ChannelDropdownMenu({
  alias,
  channel,
  hasMultipleChannels,
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
      open={!!dialog}
      onOpenChange={(open) => {
        if (!open) {
          setDialog(undefined);
        }
      }}
    >
      <DropdownMenu modal={false}>
        <Button asChild size="icon" variant="ghost">
          <DropdownMenuTrigger>
            <MoreHorizontalIcon />
          </DropdownMenuTrigger>
        </Button>
        <DropdownMenuContent align="end" className="w-64">
          {channel.status == "online" &&
            channel.remoteBalance > channel.localSpendableBalance &&
            hasMultipleChannels && (
              <AlertDialogTrigger asChild>
                <DropdownMenuItem onClick={() => setDialog("rebalance")}>
                  <ScaleIcon />
                  Rebalance In
                </DropdownMenuItem>
              </AlertDialogTrigger>
            )}
          <DropdownMenuItem>
            <ExternalLink
              className="flex flex-1 flex-row items-center gap-2"
              to={`${info?.mempoolUrl}/tx/${channel.fundingTxId}#flow=&vout=${channel.fundingTxVout}`}
            >
              <ExternalLinkIcon />
              View Funding Transaction
            </ExternalLink>
          </DropdownMenuItem>
          <DropdownMenuItem>
            <ExternalLink
              className="flex flex-1 flex-row items-center gap-2"
              to={`https://amboss.space/node/${channel.remotePubkey}`}
            >
              <ExternalLinkIcon />
              View Node on amboss.space
            </ExternalLink>
          </DropdownMenuItem>
          {channel.public && (
            <AlertDialogTrigger asChild>
              <DropdownMenuItem onClick={() => setDialog("routingFee")}>
                <HandCoinsIcon />
                Set Routing Fee
              </DropdownMenuItem>
            </AlertDialogTrigger>
          )}
          <AlertDialogTrigger asChild>
            <DropdownMenuItem onClick={() => setDialog("closeChannel")}>
              <Trash2Icon className="text-destructive" />
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
