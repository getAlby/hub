import { AlertCircleIcon } from "lucide-react";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { App } from "src/types";

export function ConnectionDetailsModal({
  app,
  onClose,
}: {
  app: App;
  onClose: () => void;
}) {
  return (
    <AlertDialog open>
      <AlertDialogContent>
        <AlertDialogTitle>Connection Details</AlertDialogTitle>
        <AlertDialogDescription className="flex flex-col gap-2">
          <div>
            <p className="font-medium">App ID</p>
            <p className="text-muted-foreground break-all">{app.id}</p>
          </div>
          <div>
            <p className="font-medium">App Public Key</p>
            <p className="text-muted-foreground break-all">{app.appPubkey}</p>
          </div>
          <div>
            <p className="font-medium">Wallet Public Key</p>
            <p className="text-muted-foreground break-all">
              <span className="break-all">{app.walletPubkey}</span>
              {!app.uniqueWalletPubkey && (
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <AlertCircleIcon className="w-3 h-3 ml-2 flex-shrink-0" />
                    </TooltipTrigger>
                    <TooltipContent className="w-[300px]">
                      This connection does not have its own unique wallet
                      pubkey. Re-connect for additional privacy.
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
            </p>
          </div>
          <div>
            <p className="font-medium">Last used</p>
            <p className="text-muted-foreground break-all">
              {app.lastUsedAt ? new Date(app.lastUsedAt).toString() : "Never"}
            </p>
          </div>

          <div>
            <p className="font-medium">Created At</p>
            <p className="text-muted-foreground break-all">
              {new Date(app.createdAt).toString()}
            </p>
          </div>

          {app.metadata && (
            <div>
              <p className="font-medium">Metadata</p>
              <p className="text-muted-foreground break-all whitespace-pre-wrap bg-neutral-100 p-2">
                {JSON.stringify(app.metadata, null, 4)}
              </p>
            </div>
          )}
        </AlertDialogDescription>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onClose}>Close</AlertDialogCancel>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
