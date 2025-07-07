import React from "react";
import { useToast } from "src/components/ui/use-toast";
import { useApp } from "src/hooks/useApp";
import { request } from "src/utils/request";

export function useCreateLightningAddress(appPubkey?: string) {
  const { toast } = useToast();
  const { data: app, mutate: refetchApp } = useApp(appPubkey);
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
      toast({
        title: "Successfully created lightning address",
      });
    } catch (error) {
      toast({
        title: "Failed to create lightning address",
        description: (error as Error).message.replace(
          "500 ",
          ""
        ) /* remove 500 error code */,
        variant: "destructive",
      });
    }
    setCreatingLightningAddress(false);
  }
  return { createLightningAddress, creatingLightningAddress };
}
