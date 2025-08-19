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

export function DisconnectApp({ app }: { app: App }) {
  const navigate = useNavigate();

  const { deleteApp, isDeleting } = useDeleteApp(() => {
    navigate(
      app.metadata?.app_store_app_id !== SUBWALLET_APPSTORE_APP_ID
        ? "/apps?tab=connected-apps"
        : "/sub-wallets"
    );
  });

  return (
    <AlertDialog>
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
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => deleteApp(app.appPubkey)}
            disabled={isDeleting}
          >
            Continue
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
