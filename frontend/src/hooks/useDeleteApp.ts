import React from "react";
import { useToast } from "src/components/ui/use-toast";

import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function useDeleteApp(onSuccess?: (nostrPubkey: string) => void) {
  const [isDeleting, setDeleting] = React.useState(false);
  const { toast } = useToast();

  const deleteApp = React.useCallback(
    async (nostrPubkey: string) => {
      setDeleting(true);
      try {
        await request(`/api/apps/${nostrPubkey}`, {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
        });
        toast({ title: "App deleted" });
        if (onSuccess) {
          onSuccess(nostrPubkey);
        }
      } catch (error) {
        await handleRequestError(toast, "Failed to delete app", error);
      } finally {
        setDeleting(false);
      }
    },
    [onSuccess, toast]
  );

  return React.useMemo(
    () => ({ deleteApp, isDeleting }),
    [deleteApp, isDeleting]
  );
}
