import React from "react";
import { useNavigate } from "react-router-dom";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

export function useUnlinkAlbyAccount(
  navigateTo = "/",
  successMessage = "Your hub is no longer connected to an Alby Account."
) {
  const { toast } = useToast();
  const navigate = useNavigate();
  const { mutate: refetchInfo } = useInfo();

  const disconnect = React.useCallback(async () => {
    try {
      await request("/api/alby/unlink-account", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });
      await refetchInfo();
      navigate(navigateTo);
      toast({
        title: "Alby Account Disconnected",
        description: successMessage,
      });
    } catch (error) {
      toast({
        title: "Disconnect account failed",
        description: (error as Error).message,
        variant: "destructive",
      });
    }
  }, [refetchInfo, navigate, navigateTo, toast, successMessage]);
  return disconnect;
}
