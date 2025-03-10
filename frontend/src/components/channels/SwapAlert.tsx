import { ArrowDownUp, ExternalLinkIcon } from "lucide-react";
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
      <AlertTitle className="flex items-center gap-1">
        <ArrowDownUp className="h-4 w-4" />
        Swap funds in or out of existing channels
      </AlertTitle>
      <AlertDescription className="text-xs text-muted-foreground">
        <p>
          It can be more economic to swap funds in and out of existing channels
          rather than opening new channels or closing existing ones.
        </p>
        <div className="flex items-center justify-end mt-2 gap-2">
          <ExternalLinkButton
            to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/node/swaps-in-and-out"
            variant="outline"
          >
            Learn more
            <ExternalLinkIcon className="w-4 h-4 ml-2" />
          </ExternalLinkButton>
          <LinkButton to="/channels?swap=true" variant="secondary">
            Swap
          </LinkButton>
        </div>
      </AlertDescription>
    </Alert>
  );
}
