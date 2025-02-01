import { AlertCircleIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { RecommendedChannelPeer } from "src/types";

type ChannelPeerNoteProps = {
  peer: RecommendedChannelPeer;
};

export function ChannelPeerNote({ peer }: ChannelPeerNoteProps) {
  return (
    <Alert>
      <AlertCircleIcon className="h-4 w-4" />
      <AlertTitle>
        Please note when opening a channel with {peer.name}
      </AlertTitle>
      <AlertDescription>{peer.note}</AlertDescription>
    </Alert>
  );
}
