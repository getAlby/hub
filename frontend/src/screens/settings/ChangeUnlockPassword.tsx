import { TriangleAlertIcon } from "lucide-react";
import React from "react";
import PasswordInput from "src/components/password/PasswordInput";

import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

export function ChangeUnlockPassword() {
  const { toast } = useToast();
  const { mutate: refetchInfo } = useInfo();

  const [currentUnlockPassword, setCurrentUnlockPassword] = React.useState("");
  const [newUnlockPassword, setNewUnlockPassword] = React.useState("");
  const [confirmNewUnlockPassword, setConfirmNewUnlockPassword] =
    React.useState("");

  const [loading, setLoading] = React.useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (newUnlockPassword !== confirmNewUnlockPassword) {
        throw new Error("Password confirmation does not match");
      }
      setLoading(true);
      await request("/api/unlock-password", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          currentUnlockPassword,
          newUnlockPassword,
        }),
      });
      await refetchInfo();
      toast({
        title: "Successfully changed password",
        description: "Please start your node with your new password",
      });
    } catch (error) {
      toast({
        title: "Password change failed",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <SettingsHeader
        title="Unlock Password"
        description="Change unlock password to your Hub. Your node will restart after password change."
      />
      <div>
        <Alert variant={"destructive"} className="w-full md:max-w-6xl mb-8">
          <AlertTitle>
            <div className="flex gap-2">
              <TriangleAlertIcon className="size-4" /> Important!
            </div>
          </AlertTitle>
          <AlertDescription>
            Password can't be reset or recovered. Make sure to back it up!
          </AlertDescription>
        </Alert>
        <form
          onSubmit={onSubmit}
          className="w-full md:w-96 flex flex-col gap-6"
        >
          <div className="grid gap-1.5">
            <Label htmlFor="current-password">Current Password</Label>
            <PasswordInput
              id="current-password"
              autoFocus
              onChange={setCurrentUnlockPassword}
              value={currentUnlockPassword}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="new-password">New Password</Label>
            <PasswordInput
              id="new-password"
              onChange={setNewUnlockPassword}
              value={newUnlockPassword}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="confirm-new-password">Confirm New Password</Label>
            <PasswordInput
              id="confirm-new-password"
              onChange={setConfirmNewUnlockPassword}
              value={confirmNewUnlockPassword}
            />
          </div>
          <div className="flex justify-start">
            <LoadingButton
              loading={loading}
              disabled={
                !(
                  currentUnlockPassword &&
                  newUnlockPassword &&
                  confirmNewUnlockPassword
                )
              }
            >
              Change Password
            </LoadingButton>
          </div>
        </form>
      </div>
    </>
  );
}
