import { Link } from "react-router-dom";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useChannels } from "src/hooks/useChannels";
import { usePeers } from "src/hooks/usePeers";

export function NodeStatusWidget() {
  const { data: channels } = useChannels();
  const { data: peers } = usePeers();

  if (!channels || !peers) {
    return null;
  }
  return (
    <Link to="/channels">
      <Card>
        <CardHeader>
          <CardTitle>Node</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-xs">Channels Online</p>
          <p className="text-xl font-semibold">
            {channels.filter((c) => c.active).length || 0} /{" "}
            {channels.filter((c) => c.active).length || 0}
          </p>
          <p className="text-muted-foreground text-xs mt-6">Connected Peers</p>
          <p className="text-xl font-semibold">
            {peers.filter((p) => p.isConnected).length || 0} /{" "}
            {peers.filter((p) => p.isConnected).length || 0}
          </p>
        </CardContent>
      </Card>
    </Link>
  );
}
