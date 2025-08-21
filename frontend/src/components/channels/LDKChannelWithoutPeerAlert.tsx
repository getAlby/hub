import { AlertTriangleIcon, ExternalLinkIcon } from "lucide-react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { usePeers } from "src/hooks/usePeers";
import { Channel } from "src/types";

export function LDKChannelWithoutPeerAlert() {
  const { data: info } = useInfo();
  const { data: peers } = usePeers();
  const { data: channels } = useChannels();

  if (info?.backendType !== "LDK") {
    // LND does not show disconnected peers
    return null;
  }

  const channelWithoutPeer = channels?.find(
    (channel) => !peers?.some((peer) => peer.nodeId === channel.remotePubkey)
  );
  if (!channelWithoutPeer) {
    return null;
  }

  return <ChannelWithoutPeerAlertInternal channel={channelWithoutPeer} />;
}

function ChannelWithoutPeerAlertInternal({ channel }: { channel: Channel }) {
  const { data: nodeDetails } = useNodeDetails(channel.remotePubkey);
  return (
    <Alert>
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>
        Channel with peer{" "}
        {nodeDetails?.alias || channel.remotePubkey.substring(0, 8) + "..."} is
        not peered
      </AlertTitle>
      <AlertDescription className="gap-0">
        <p>
          Your channel will be offline until you manually reconnect to this
          peer.
        </p>
        <p>
          1. Copy the peer connection details from{" "}
          <ExternalLink
            to={`https://amboss.space/node/${channel.remotePubkey}`}
            className="font-semibold inline-flex gap-1"
          >
            <ExternalLinkIcon className="size-4" />
            amboss.space
          </ExternalLink>{" "}
        </p>
        <p>
          2. Open{" "}
          <Link to="/peers" className="font-semibold">
            Peers
          </Link>{" "}
          to re-connect the peer.
        </p>
      </AlertDescription>
    </Alert>
  );
}
