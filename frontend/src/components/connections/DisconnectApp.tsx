import { UnplugIcon } from "lucide-react";
import { useNavigate } from "react-router";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { ROUTSTR_APP_STORE_ID } from "src/components/connections/routstr/constants";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useDeleteApp } from "src/hooks/useDeleteApp";
import { App } from "src/types";

export function DisconnectApp({
  app,
  onClose,
}: {
  app: App;
  onClose: () => void;
}) {
  const navigate = useNavigate();

  const { deleteApp, isDeleting } = useDeleteApp(app, () => {
    navigate(
      app.metadata?.app_store_app_id !== SUBWALLET_APPSTORE_APP_ID
        ? "/apps?tab=connected-apps"
        : "/sub-wallets"
    );
  });

  const isSubwallet =
    app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID;
  const hasLightningAddress = !!app.metadata?.lud16;

  const isRoutstr = app.metadata?.app_store_app_id === ROUTSTR_APP_STORE_ID;
  const routstrMeta = (app.metadata as Record<string, unknown> | undefined)
    ?.routstr as { apiKey?: string } | undefined;
  const hasRoutstrKey = !!(isRoutstr && routstrMeta?.apiKey);

  return (
    <AlertDialog
      open
      onOpenChange={(open) => {
        if (!open) {
          onClose();
        }
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {hasRoutstrKey
              ? "Delete API key before disconnecting"
              : "Are you sure you want to delete this connection?"}
          </AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3 text-sm text-muted-foreground">
              {hasRoutstrKey ? (
                <>
                  <p>
                    This Routstr connection still has an API key. Disconnecting
                    now would leave the key orphaned in the local daemon.
                  </p>
                  <ol className="list-decimal list-inside space-y-1">
                    <li>Refund any remaining balance to this Routstr wallet</li>
                    <li>Delete the API key</li>
                    <li>Then disconnect this connection</li>
                  </ol>
                  <p>
                    Use Refund and Delete Key on this page, then try disconnect
                    again.
                  </p>
                </>
              ) : (
                <>
                  <p>
                    Connected apps will no longer be able to use this
                    connection.
                    {app.isolated && (
                      <>
                        {" "}
                        No funds will be lost during this process, the balance
                        will remain in your wallet.
                      </>
                    )}
                  </p>
                  {isSubwallet && hasLightningAddress && (
                    <p className="font-medium text-foreground">
                      This sub-wallet has a lightning address (
                      {String(app.metadata?.lud16)}) that will also be deleted.
                    </p>
                  )}
                  {isRoutstr && app.isolated && (
                    <p>
                      Optional: decrease any remaining isolated balance back to
                      your main wallet before disconnecting.
                    </p>
                  )}
                </>
              )}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onClose}>
            {hasRoutstrKey ? "Close" : "Cancel"}
          </AlertDialogCancel>
          {!hasRoutstrKey && (
            <AlertDialogAction onClick={deleteApp} disabled={isDeleting}>
              <UnplugIcon className="size-4" />
              Confirm
            </AlertDialogAction>
          )}
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
