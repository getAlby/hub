import { TriangleAlertIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useChannels } from "src/hooks/useChannels";

type PeerAlertProps = {
  pubkey?: string;
  name?: string;
};

export function DuplicateChannelAlert({ pubkey, name }: PeerAlertProps) {
  const { data: channels } = useChannels();

  if (!pubkey) {
    return null;
  }

  const matchedPeer = channels?.find((p) => p.remotePubkey === pubkey);

  if (!matchedPeer) {
    return null;
  }

  return (
    <Alert>
      <TriangleAlertIcon />
      <AlertTitle>
        You already have a channel with{" "}
        {name && name !== "Custom" ? (
          <span className="font-semibold">{name}</span>
        ) : (
          "the selected peer"
        )}
      </AlertTitle>
      <AlertDescription>
        There are other options available rather than opening multiple channels
        with the same counterparty.{" "}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/alby-hub/faq/should-i-open-multiple-channels-with-the-same-counterparty"
          className="underline"
        >
          Learn more
        </ExternalLink>
      </AlertDescription>
    </Alert>
  );
}
