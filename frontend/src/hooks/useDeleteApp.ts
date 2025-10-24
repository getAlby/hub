import React from "react";
import { toast } from "sonner";

import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useApps } from "src/hooks/useApps";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function useDeleteApp(onSuccess?: (appPubkey: string) => void) {
  const [isDeleting, setDeleting] = React.useState(false);
  const { data: appsData } = useApps();

  const deleteApp = React.useCallback(
    async (appPubkey: string) => {
      setDeleting(true);
      try {
        // Find the app to check if it's a sub-wallet with a lightning address
        const app = appsData?.apps.find((a) => a.appPubkey === appPubkey);
        const isSubwallet =
          app?.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID;
        const hasLightningAddress = !!app?.metadata?.lud16;

        // Delete lightning address first if it exists for a sub-wallet
        if (isSubwallet && hasLightningAddress && app?.id) {
          await request(`/api/lightning-addresses/${app.id}`, {
            method: "DELETE",
            headers: {
              "Content-Type": "application/json",
            },
          });
        }

        // Delete the app/sub-wallet
        await request(`/api/apps/${appPubkey}`, {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
        });

        toast("Connection deleted");

        if (onSuccess) {
          onSuccess(appPubkey);
        }
      } catch (error) {
        await handleRequestError("Failed to delete connection", error);
      } finally {
        setDeleting(false);
      }
    },
    [onSuccess, appsData?.apps]
  );

  return React.useMemo(
    () => ({ deleteApp, isDeleting }),
    [deleteApp, isDeleting]
  );
}
