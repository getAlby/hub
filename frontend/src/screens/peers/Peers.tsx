import { MoreHorizontal, Trash2 } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader.tsx";
import { DisconnectPeerDialog } from "src/components/DisconnectPeerDialog.tsx";
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
import { toast } from "src/components/ui/use-toast";
import { useChannels } from "src/hooks/useChannels";
import { usePeers } from "src/hooks/usePeers.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { Node, Peer } from "src/types";
import { request } from "src/utils/request";

export default function Peers() {
  useSyncWallet();
  const { data: peers } = usePeers();
  const { data: channels } = useChannels();
  const [nodes, setNodes] = React.useState<Node[]>([]);
  const [peer, setPeer] = React.useState<Peer>();

  // TODO: move to NWC backend
  const loadNodeStats = React.useCallback(async () => {
    if (!peers) {
      return [];
    }
    const nodes = await Promise.all(
      peers?.map(async (peer): Promise<Node | undefined> => {
        try {
          const response = await request<Node>(
            `/api/mempool?endpoint=/v1/lightning/nodes/${peer.nodeId}`
          );
          return response;
        } catch (error) {
          console.error(error);
          return undefined;
        }
      })
    );
    setNodes(nodes.filter((node) => !!node) as Node[]);
  }, [peers]);

  React.useEffect(() => {
    loadNodeStats();
  }, [loadNodeStats]);

  async function checkDisconnectPeer(peer: Peer) {
    try {
      if (!channels) {
        throw new Error("channels not loaded");
      }
      if (channels.some((channel) => channel.remotePubkey === peer.nodeId)) {
        throw new Error(
          "you have one or more open channels with " + peer.nodeId
        );
      }
      setPeer(peer);
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to disconnect peer: " + e,
      });
      console.error(e);
    }
  }

  return (
    <>
      <AppHeader
        title="Peers"
        description="Manage your connections with other lightning nodes"
        contentRight={
          <>
            <Link to="/peers/new">
              <Button>Connect Peer</Button>
            </Link>
          </>
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
          <>
            {peers?.map((peer) => {
              const node = nodes.find((n) => n.public_key === peer.nodeId);
              const alias = node?.alias || "Unknown";

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
                        {alias}
                      </Button>
                    </a>
                  </TableCell>
                  <TableCell>{peer.nodeId}</TableCell>
                  <TableCell>{peer.address}</TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button size="icon" variant="ghost">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => checkDisconnectPeer(peer)}
                          className="flex flex-row items-center gap-2"
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                          Disconnect Peer
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              );
            })}
          </>
        </TableBody>
      </Table>

      <AlertDialog
        open={!!peer}
        onOpenChange={(open) => {
          if (!open) {
            setPeer(undefined);
          }
        }}
      >
        {peer && <DisconnectPeerDialog peer={peer} />}
      </AlertDialog>
    </>
  );
}
