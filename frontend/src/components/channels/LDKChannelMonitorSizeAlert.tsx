import { AlertTriangleIcon } from "lucide-react";
import React from "react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useInfo } from "src/hooks/useInfo";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { request } from "src/utils/request";

export function LDKChannelMonitorSizeAlert() {
  const { data: info } = useInfo();

  if (info?.backendType !== "LDK") {
    return null;
  }
  return ChannelMonitorSizeAlert();
}

function ChannelMonitorSizeAlert() {
  const [channelMonitorSizes, setChannelMonitorSizes] =
    React.useState<
      { remotePubkey: string; sizeBytes: number; hasWarning: boolean }[]
    >();
  React.useEffect(() => {
    (async () => {
      try {
        const requestOptions: RequestInit = {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ command: "list_channel_monitor_sizes" }),
        };

        const data = await request("/api/command", requestOptions);

        setChannelMonitorSizes(data as typeof channelMonitorSizes);
      } catch (error) {
        console.error(error);
      }
    })();
  }, []);

  if (!channelMonitorSizes) {
    return null;
  }

  const largestChannelMonitor = channelMonitorSizes.find(
    (c1) => !channelMonitorSizes.find((c2) => c2.sizeBytes > c1.sizeBytes)
  );
  if (!largestChannelMonitor) {
    return null;
  }

  if (!largestChannelMonitor.hasWarning) {
    return;
  }
  return (
    <ChannelMonitorSizeAlertForPubkey
      remotePubkey={largestChannelMonitor.remotePubkey}
      sizeBytes={largestChannelMonitor.sizeBytes}
    />
  );
}

function ChannelMonitorSizeAlertForPubkey({
  remotePubkey,
  sizeBytes,
}: {
  remotePubkey: string;
  sizeBytes: number;
}) {
  const { data: peerDetails } = useNodeDetails(remotePubkey);
  return (
    <>
      <Alert>
        <AlertTriangleIcon className="h-4 w-4" />
        <AlertTitle>Large channel state detected</AlertTitle>
        <AlertDescription>
          The channel state for your channel with{" "}
          {peerDetails?.alias || remotePubkey} is over{" "}
          {Math.floor(sizeBytes / 1_000_000)} MB. Consider closing this channel
          and opening a new one to improve your node performance.
        </AlertDescription>
      </Alert>
    </>
  );
}
