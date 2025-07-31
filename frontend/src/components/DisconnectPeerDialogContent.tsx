import { useToast } from "src/components/ui/use-toast";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { usePeers } from "src/hooks/usePeers";
import { Peer } from "src/types";
import { request } from "src/utils/request";
import {
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";

type Props = {
  peer: Peer;
};

export function DisconnectPeerDialogContent({ peer }: Props) {
  const { mutate: reloadPeers } = usePeers();
  const { data: peerDetails } = useNodeDetails(peer.nodeId);
  const { toast } = useToast();

  async function disconnectPeer() {
    try {
      console.info(`Disconnecting from ${peer.nodeId}`);

      await request(`/api/peers/${peer.nodeId}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      });
      toast({
        title: "Successfully disconnected from peer",
        description: peer.nodeId,
      });
      await reloadPeers();
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to disconnect peer",
        description: "" + e,
      });
      console.error(e);
    }
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Disconnect Peer</AlertDialogTitle>
        <AlertDialogDescription>
          <div>
            <p>
              Are you sure you wish to disconnect from{" "}
              {peerDetails?.alias || "this peer"}?
            </p>
            <p className="text-primary font-medium mt-4">Peer Pubkey</p>
            <p className="break-all">{peer.nodeId}</p>
          </div>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction onClick={disconnectPeer}>Confirm</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}
