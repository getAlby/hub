import React from "react";
import { toast } from "sonner";

import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { App } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function useDeleteApp(app: App, onSuccess?: () => void) {
  const [isDeleting, setDeleting] = React.useState(false);

  const deleteApp = React.useCallback(async () => {
    setDeleting(true);
    try {
      // Find the app to check if it's a sub-wallet with a lightning address
      const isSubwallet =
        app?.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID;
      const hasLightningAddress = !!app?.metadata?.lud16;

      // Delete lightning address first if it exists for a sub-wallet
      if (isSubwallet && hasLightningAddress && app?.id) {
        try {
          await request(`/api/lightning-addresses/${app.id}`, {
            method: "DELETE",
            headers: {
              "Content-Type": "application/json",
            },
          });
        } catch (error) {
          console.error("failed to delete sub-wallet lightning address", error);
          toast.error("failed to delete sub-wallet lightning address", {
            description: "" + error,
          });
        }
      }

      // Delete the app/sub-wallet
      await request(`/api/apps/${app.appPubkey}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      });

      toast("Connection deleted");

      if (onSuccess) {
        onSuccess();
      }
    } catch (error) {
      await handleRequestError("Failed to delete connection", error);
    } finally {
      setDeleting(false);
    }
  }, [onSuccess, app]);

  return React.useMemo(
    () => ({ deleteApp, isDeleting }),
    [deleteApp, isDeleting]
  );
}
