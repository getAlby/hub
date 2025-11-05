import React from "react";
import { toast } from "sonner";

import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function useDeleteApp(onSuccess?: (appPubkey: string) => void) {
  const [isDeleting, setDeleting] = React.useState(false);

  const deleteApp = React.useCallback(
    async (appPubkey: string) => {
      setDeleting(true);
      try {
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
    [onSuccess]
  );

  return React.useMemo(
    () => ({ deleteApp, isDeleting }),
    [deleteApp, isDeleting]
  );
}
