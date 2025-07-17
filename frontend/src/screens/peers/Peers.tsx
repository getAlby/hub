import { MoreHorizontalIcon, Trash2Icon } from "lucide-react";
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
import { useToast } from "src/components/ui/use-toast";
import { usePeers } from "src/hooks/usePeers.ts";
import { useSyncWallet } from "src/hooks/useSyncWallet.ts";
import { Peer } from "src/types";

export default function Peers() {
  useSyncWallet();
  const { data: peers } = usePeers();
  const { toast } = useToast();
  const [peerToDisconnect, setPeerToDisconnect] = React.useState<Peer>();

  async function checkDisconnectPeer(peer: Peer) {
    try {
      if (peer.hasOpenedChannel) {
        throw new Error(
          "you have one or more open channels with " + peer.nodeId
        );
      }
      setPeerToDisconnect(peer);
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Cannot disconnect peer",
        description: "" + e,
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
              const alias = peer.nodeAlias || "Unknown";

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
                    <DropdownMenu modal={false}>
                      <DropdownMenuTrigger asChild>
                        <Button size="icon" variant="ghost">
                          <MoreHorizontalIcon className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => checkDisconnectPeer(peer)}
                          className="flex flex-row items-center gap-2"
                        >
                          <Trash2Icon className="h-4 w-4 text-destructive" />
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
        open={!!peerToDisconnect}
        onOpenChange={(open) => {
          if (!open) {
            setPeerToDisconnect(undefined);
          }
        }}
      >
        {peerToDisconnect && (
          <DisconnectPeerDialogContent
            peer={peerToDisconnect}
            name={peerToDisconnect.nodeAlias}
          />
        )}
      </AlertDialog>
    </>
  );
}
