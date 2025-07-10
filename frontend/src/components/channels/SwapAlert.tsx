import { ArrowDownUpIcon, ExternalLinkIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { ExternalLinkButton, LinkButton } from "src/components/ui/button";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

type SwapAlertProps = {
  className?: string;
  minChannels?: number;
};
export function SwapAlert({ className, minChannels = 2 }: SwapAlertProps) {
  const { data: channels } = useChannels();
  const { data: balances } = useBalances();

  if (!channels || !balances) {
    return null;
  }
  if (minChannels && channels.length < minChannels) {
    return null;
  }

  const isSwapOut =
    balances.lightning.totalSpendable > balances.lightning.totalReceivable;
  const directionText = isSwapOut ? "out from" : "into";

  return (
    <Alert className={className}>
      <AlertTitle className="flex items-center gap-1">
        <ArrowDownUpIcon className="h-4 w-4" />
        Swap {directionText} existing channels
      </AlertTitle>
      <AlertDescription className="text-xs text-muted-foreground">
        <p>
          It can be more economic to swap funds {directionText} existing
          channels rather than opening new channels or closing existing ones.
        </p>
        <div className="flex items-center justify-end mt-2 gap-2">
          <ExternalLinkButton
            to="https://guides.getalby.com/user-guide/alby-hub/node/swap-in-and-out"
            variant="outline"
          >
            Learn more
            <ExternalLinkIcon className="size-4 ml-2" />
          </ExternalLinkButton>
          <LinkButton to="/channels?swap=true" variant="secondary">
            Swap {isSwapOut ? "Out" : "In"}
          </LinkButton>
        </div>
      </AlertDescription>
    </Alert>
  );
}
