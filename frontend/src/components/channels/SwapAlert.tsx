import { RefreshCcw } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { ExternalLinkButton, LinkButton } from "src/components/ui/button";
import { useChannels } from "src/hooks/useChannels";

type SwapAlertProps = {
  className?: string;
};
export function SwapAlert({ className }: SwapAlertProps) {
  const { data: channels } = useChannels();

  if (!channels) {
    return null;
  }
  if (channels.length < 2) {
    return null;
  }

  return (
    <Alert className={className}>
      <AlertTitle className="flex items-center gap-2">
        <RefreshCcw className="h-4 w-4" />
        Swap funds in or out of existing channels
      </AlertTitle>
      <AlertDescription>
        <p>
          It can be more economic to swap funds in and out of existing channels
          rather than opening new channels or closing existing ones.
        </p>
        <div className="flex justify-end mt-2 gap-2">
          <ExternalLinkButton
            to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/node/swaps-in-and-out"
            variant="secondary"
          >
            Learn more
          </ExternalLinkButton>
          <LinkButton to="/channels?swap=true">Swap</LinkButton>
        </div>
      </AlertDescription>
    </Alert>
  );
}
