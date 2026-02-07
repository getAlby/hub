import { CopyIcon, GlobeIcon } from "lucide-react";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { copyToClipboard } from "src/lib/clipboard";

export default function TorSettings() {
  const { data: info } = useInfo();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();

  if (!info) {
    return <Loading />;
  }

  const torAddress = nodeConnectionInfo?.torAddress;
  const torPort = nodeConnectionInfo?.torPort;
  const hasTor = !!torAddress && torPort !== undefined;
  const torUri = hasTor
    ? `${nodeConnectionInfo?.pubkey}@${torAddress}:${torPort}`
    : "";

  return (
    <>
      <SettingsHeader
        title="Tor"
        description="Tor hidden service for accepting incoming peer connections."
      />
      <div className="mt-6 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <GlobeIcon className="size-5" />
              Tor Hidden Service
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {!info.torEnabled && (
              <div className="space-y-2">
                <p className="text-muted-foreground text-sm">
                  Tor is not enabled. To enable incoming peer connections via
                  Tor, set the{" "}
                  <code className="bg-muted px-1 py-0.5 rounded text-xs">
                    LDK_TOR_ENABLED=true
                  </code>{" "}
                  environment variable and restart your node.
                </p>
                <p className="text-muted-foreground text-sm">
                  You also need a Tor daemon running with the control port
                  accessible. On Umbrel, Tor is already running. Configure the
                  control host/port with{" "}
                  <code className="bg-muted px-1 py-0.5 rounded text-xs">
                    LDK_TOR_CONTROL_HOST
                  </code>{" "}
                  and{" "}
                  <code className="bg-muted px-1 py-0.5 rounded text-xs">
                    LDK_TOR_CONTROL_PORT
                  </code>
                  .
                </p>
              </div>
            )}

            {info.torEnabled && (
              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">Status:</span>
                  {hasTor ? (
                    <Badge variant="default">Active</Badge>
                  ) : (
                    <Badge variant="secondary">Connecting...</Badge>
                  )}
                </div>

                {hasTor && (
                  <>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">
                        Onion Address
                      </label>
                      <div className="flex items-center gap-2">
                        <code className="bg-muted px-3 py-2 rounded text-xs break-all flex-1">
                          {torAddress}
                        </code>
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => copyToClipboard(torAddress)}
                        >
                          <CopyIcon className="size-4" />
                        </Button>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">
                        Node URI (Tor)
                      </label>
                      <p className="text-muted-foreground text-xs">
                        Share this URI with other Lightning nodes to let them
                        connect to you directly.
                      </p>
                      <div className="flex items-center gap-2">
                        <code className="bg-muted px-3 py-2 rounded text-xs break-all flex-1">
                          {torUri}
                        </code>
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => copyToClipboard(torUri)}
                        >
                          <CopyIcon className="size-4" />
                        </Button>
                      </div>
                    </div>
                  </>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  );
}
