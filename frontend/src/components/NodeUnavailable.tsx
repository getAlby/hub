import { RefreshCwIcon, ServerOffIcon, SettingsIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import { LinkButton } from "src/components/ui/custom/link-button";

type Props = {
  onRetry: () => void;
};

export function NodeUnavailable({ onRetry }: Props) {
  return (
    <>
      <AppHeader title="Node unavailable" pageTitle="Node unavailable" />
      <div className="flex flex-1 items-center justify-center rounded-lg bg-muted p-8">
        <div className="flex max-w-md flex-col items-center gap-1 text-center">
          <ServerOffIcon className="size-10 text-muted-foreground" />
          <h2 className="mt-4 text-lg font-semibold">
            Your lightning node is unavailable
          </h2>
          <p className="text-sm text-muted-foreground">
            Alby Hub cannot reach your lightning node right now. Check your node
            settings or try again after the node is running.
          </p>
          <div className="mt-4 flex flex-wrap justify-center gap-2">
            <Button onClick={onRetry}>
              <RefreshCwIcon />
              Try again
            </Button>
            <LinkButton to="/settings/node" variant="secondary">
              <SettingsIcon />
              Node settings
            </LinkButton>
          </div>
        </div>
      </div>
    </>
  );
}
