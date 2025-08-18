import React from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

type UnlinkAlbyAccountProps = {
  navigateTo?: string;
  successMessage?: string;
};

export function UnlinkAlbyAccount({
  children,
  navigateTo = "/",
  successMessage = "Your hub is no longer connected to an Alby Account.",
}: React.PropsWithChildren<UnlinkAlbyAccountProps>) {
  const navigate = useNavigate();
  const { mutate: refetchInfo } = useInfo();

  const unlinkAccount = React.useCallback(async () => {
    try {
      await request("/api/alby/unlink-account", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });
      await refetchInfo();
      navigate(navigateTo);
      toast("Alby Account Disconnected", {
        description: successMessage,
      });
    } catch (error) {
      toast.error("Disconnect account failed", {
        description: (error as Error).message,
      });
    }
  }, [refetchInfo, navigate, navigateTo, successMessage]);

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>{children}</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Disconnect Alby Account</AlertDialogTitle>
          <AlertDialogDescription>
            <div>
              <p>Are you sure you want to disconnect your Alby Account?</p>
              <p className="text-destructive font-medium mt-4">
                Your Alby Account will be disconnected and all Alby Account
                features such as your lightning address will stop working.
              </p>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={unlinkAccount}>Confirm</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
