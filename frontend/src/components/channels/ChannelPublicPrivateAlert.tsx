import { AlertCircleIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";

export function ChannelPublicPrivateAlert() {
  return (
    <Alert>
      <AlertCircleIcon className="h-4 w-4" />
      <AlertTitle>Conflicting Private / Public Channels</AlertTitle>
      <AlertDescription>
        <div className="mb-2">
          You will not be able to receive payments on any private channels. It
          is recommended to only open all private or all public channels.
        </div>
        <ExternalLink
          to={
            "https://guides.getalby.com/user-guide/alby-hub/faq/should-i-open-a-private-or-public-channel"
          }
          className="underline"
        >
          Learn more
        </ExternalLink>
      </AlertDescription>
    </Alert>
  );
}
