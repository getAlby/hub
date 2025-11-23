import React from "react";
import { toast } from "sonner";
import { useApp } from "src/hooks/useApp";
import { request } from "src/utils/request";

export function useCreateLightningAddress(appId?: number) {
  const { data: app, mutate: refetchApp } = useApp(appId);
  const [creatingLightningAddress, setCreatingLightningAddress] =
    React.useState(false);

  async function createLightningAddress(intendedLightningAddress: string) {
    try {
      if (!app) {
        throw new Error("app not found");
      }
      setCreatingLightningAddress(true);
      await request("/api/lightning-addresses", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          address: intendedLightningAddress,
          appId: app.id,
        }),
      });
      await refetchApp();
      toast("Successfully created lightning address");
    } catch (error) {
      toast.error("Failed to create lightning address", {
        description: (error as Error).message.replace(
          "500 ",
          ""
        ) /* remove 500 error code */,
      });
    }
    setCreatingLightningAddress(false);
  }
  return { createLightningAddress, creatingLightningAddress };
}
