import React from "react";
import { useToast } from "src/components/ui/use-toast";
import { useApp } from "src/hooks/useApp";
import { request } from "src/utils/request";

export function useDeleteLightningAddress(appId?: number) {
  const { toast } = useToast();
  const { data: app, mutate: refetchApp } = useApp(appId);
  const [deletingLightningAddress, setDeletingLightningAddress] =
    React.useState(false);

  async function deleteLightningAddress() {
    try {
      if (!app) {
        throw new Error("app not found");
      }
      setDeletingLightningAddress(true);
      await request(`/api/lightning-addresses/${app.id}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      });
      await refetchApp();
      toast({
        title: "Successfully deleted lightning address",
      });
    } catch (error) {
      toast({
        title: "Failed to delete lightning address",
        description: (error as Error).message,
        variant: "destructive",
      });
    }
    setDeletingLightningAddress(false);
  }
  return {
    deleteLightningAddress,
    deletingLightningAddress,
  };
}
