import { UnplugIcon } from "lucide-react";
import { useNavigate } from "react-router-dom";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
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

  const { deleteApp, isDeleting } = useDeleteApp(() => {
    navigate(
      app.metadata?.app_store_app_id !== SUBWALLET_APPSTORE_APP_ID
        ? "/apps?tab=connected-apps"
        : "/sub-wallets"
    );
  });

  // Check if this is a sub-wallet with a lightning address
  const isSubwallet =
    app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID;
  const hasLightningAddress = !!app.metadata?.lud16;

  return (
    <AlertDialog open>
      <AlertDialogTrigger asChild>
        <Button variant="outline">
          <UnplugIcon className="size-4" /> Disconnect
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Are you sure you want to delete this connection?
          </AlertDialogTitle>
          <AlertDialogDescription>
            Connected apps will no longer be able to use this connection.
            {app.isolated && (
              <>
                {" "}
                No funds will be lost during this process, the balance will
                remain in your wallet.
              </>
            )}
            {isSubwallet && hasLightningAddress && (
              <p className="font-medium">
                This sub-wallet has a lightning address ({app.metadata?.lud16})
                that will also be deleted.
              </p>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onClose}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => deleteApp(app.appPubkey)}
            disabled={isDeleting}
          >
            Confirm
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
