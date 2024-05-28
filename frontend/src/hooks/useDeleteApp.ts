import React from "react";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function useDeleteApp(onSuccess?: (nostrPubkey: string) => void) {
  const { data: csrf } = useCSRF();
  const [isDeleting, setDeleting] = React.useState(false);
  const { toast } = useToast();

  const deleteApp = React.useCallback(
    async (nostrPubkey: string) => {
      if (!csrf) {
        toast({ title: "No CSRF token.", variant: "destructive" });
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
        toast({ title: "App disconnected" });
        if (onSuccess) {
          onSuccess(nostrPubkey);
        }
      } catch (error) {
        await handleRequestError(toast, "Failed to delete app", error);
      } finally {
        setDeleting(false);
      }
    },
    [csrf, onSuccess, toast]
  );

  return React.useMemo(
    () => ({ deleteApp, isDeleting }),
    [deleteApp, isDeleting]
  );
}
