import React from "react";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";
import toast from "src/components/Toast";
import { handleRequestError } from "src/utils/handleRequestError";

export function useDeleteApp(onSuccess?: (nostrPubkey: string) => void) {
  const { data: csrf } = useCSRF();
  const [isDeleting, setDeleting] = React.useState(false);

  const deleteApp = React.useCallback(
    async (nostrPubkey: string) => {
      if (!csrf) {
        toast.error("No CSRF token");
        return;
      }

      setDeleting(true);
      try {
        await request(`/api/apps/${nostrPubkey}`, {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": csrf,
          },
        });
        toast.success("App disconnected");
        if (onSuccess) {
          onSuccess(nostrPubkey);
        }
      } catch (error) {
        await handleRequestError("Failed to delete app", error);
      } finally {
        setDeleting(false);
      }
    },
    [csrf, onSuccess]
  );

  return React.useMemo(
    () => ({ deleteApp, isDeleting }),
    [deleteApp, isDeleting]
  );
}
