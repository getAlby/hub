import { MoreHorizontalIcon, PlugZapIcon, Trash2Icon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader.tsx";
import { DisconnectPeerDialogContent } from "src/components/DisconnectPeerDialogContent";
import { AlertDialog } from "src/components/ui/alert-dialog.tsx";
import { Badge } from "src/components/ui/badge.tsx";
import { Button } from "src/components/ui/button.tsx";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu.tsx";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "src/components/ui/table.tsx";
import { useChannels } from "src/hooks/useChannels";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { usePeers } from "src/hooks/usePeers.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { splitSocketAddress } from "src/lib/utils";
import { ConnectPeerRequest, Peer } from "src/types";
import { request } from "src/utils/request";

export default function Peers() {
  useSyncWallet();
  const { data: peers } = usePeers();
  const [peerToDisconnect, setPeerToDisconnect] = React.useState<Peer>();

  return (
    <>
      <AppHeader
        title="Peers"
        description="Manage your connections with other lightning nodes"
        contentRight={
          <Link to="/peers/new">
            <Button>Connect Peer</Button>
          </Link>
        }
      ></AppHeader>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[80px]">Status</TableHead>
            <TableHead>Node</TableHead>
            <TableHead>Pubkey</TableHead>
            <TableHead>Address</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {peers?.map((peer) => (
            <PeerTableRow
              key={peer.nodeId}
              peer={peer}
              setPeerToDisconnect={setPeerToDisconnect}
            />
          ))}
        </TableBody>
      </Table>

      <AlertDialog
        open={!!peerToDisconnect}
        onOpenChange={(open) => {
          if (!open) {
            setPeerToDisconnect(undefined);
          }
        }}
      >
        {peerToDisconnect && (
          <DisconnectPeerDialogContent peer={peerToDisconnect} />
        )}
      </AlertDialog>
    </>
  );
}

type PeerTableRowProps = {
  peer: Peer;
  setPeerToDisconnect: (peer: Peer) => void;
};

function PeerTableRow(props: PeerTableRowProps) {
  const { peer } = props;
  const { data: channels } = useChannels();
  const { data: peerDetails } = useNodeDetails(peer.nodeId);

  function hasOpenedChannels(peer: Peer) {
    return channels?.some((channel) => channel.remotePubkey === peer.nodeId);
  }

  const connectPeer = async (peer: Peer) => {
    const { address, port } = splitSocketAddress(peer.address);

    if (!address || !port) {
      throw new Error("host not found");
    }
    console.info(`🔌 Peering with ${peer.nodeId}`);
    const connectPeerRequest: ConnectPeerRequest = {
      pubkey: peer.nodeId,
      address,
      port: +port,
    };
    await request("/api/peers", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(connectPeerRequest),
    });
  };

  return (
    <TableRow key={peer.nodeId}>
      <TableCell>
        {peer.isConnected ? (
          <Badge>Online</Badge>
        ) : (
          <Badge variant="outline">Offline</Badge>
        )}{" "}
      </TableCell>
      <TableCell className="flex flex-row items-center">
        <a
          title={peer.nodeId}
          href={`https://amboss.space/node/${peer.nodeId}`}
          target="_blank"
          rel="noopener noreferer"
        >
          <Button variant="link" className="p-0 mr-2">
            {peerDetails?.alias}
          </Button>
        </a>
      </TableCell>
      <TableCell>{peer.nodeId}</TableCell>
      <TableCell>{peer.address}</TableCell>
      <TableCell>
        {(!hasOpenedChannels(peer) || !peer.isConnected) && (
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger>
              <Button size="icon" variant="ghost">
                <MoreHorizontalIcon className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            {channels && (
              <DropdownMenuContent align="end">
                {!hasOpenedChannels(peer) && (
                  <DropdownMenuItem
                    onClick={() => props.setPeerToDisconnect(peer)}
                    className="flex flex-row items-center gap-2"
                  >
                    <Trash2Icon className="size-4 text-destructive" />
                    Disconnect Peer
                  </DropdownMenuItem>
                )}
                {!peer.isConnected && (
                  <DropdownMenuItem
                    onClick={() => connectPeer(peer)}
                    className="flex flex-row items-center gap-2"
                  >
                    <PlugZapIcon className="size-4" />
                    Try to Reconnect
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            )}
          </DropdownMenu>
        )}
      </TableCell>
    </TableRow>
  );
}
