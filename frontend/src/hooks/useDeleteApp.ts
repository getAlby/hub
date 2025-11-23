import React from "react";
import { toast } from "sonner";

import { App } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function useDeleteApp(app: App, onSuccess?: () => void) {
  const [isDeleting, setDeleting] = React.useState(false);

  const deleteApp = React.useCallback(async () => {
    setDeleting(true);
    try {
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
