import { MoreHorizontal, Trash2 } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader.tsx";
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
import { toast } from "src/components/ui/use-toast.ts";
import { useChannels } from "src/hooks/useChannels";
import { usePeers } from "src/hooks/usePeers.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { Node } from "src/types";
import { request } from "src/utils/request";
import { useCSRF } from "../../hooks/useCSRF.ts";

export default function Peers() {
  useSyncWallet();
  const { data: peers, mutate: reloadPeers } = usePeers();
  const { data: channels } = useChannels();
  const [nodes, setNodes] = React.useState<Node[]>([]);
  const { data: csrf } = useCSRF();

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

  async function disconnectPeer(peerId: string) {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      if (!peerId) {
        throw new Error("peer missing");
      }
      if (!channels) {
        throw new Error("channels not loaded");
      }
      if (channels.some((channel) => channel.remotePubkey === peerId)) {
        throw new Error("you have one or more open channels with " + peerId);
      }
      if (
        !confirm(
          "Are you sure you wish to disconnect with peer " + peerId + "?"
        )
      ) {
        return;
      }
      console.log(`Disconnecting from ${peerId}`);

      await request(`/api/peers/${peerId}`, {
        method: "DELETE",
        headers: {
          "X-CSRF-Token": csrf,
        },
      });
      toast({ title: "Successfully disconnected from peer " + peerId });
      await reloadPeers();
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
                          className="flex flex-row items-center gap-2"
                          onClick={() => disconnectPeer(peer.nodeId)}
                        >
                          <Trash2 className="text-destructive" />
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
    </>
  );
}
