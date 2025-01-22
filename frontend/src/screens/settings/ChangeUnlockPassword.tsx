import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import React from "react";
import PasswordViewAdornment from "src/components/PasswordAdornment";

import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Input } from "src/components/ui/input";
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
  const [currentUnlockPasswordVisible, setCurrentUnlockPasswordVisible] =
    React.useState(false);
  const [newUnlockPasswordVisible, setNewUnlockPasswordVisible] =
    React.useState(false);
  const [confirmNewUnlockPasswordVisible, setConfirmNewUnlockPasswordVisible] =
    React.useState(false);

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
        <Alert variant={"destructive"} className="mb-8">
          <AlertTitle>
            <div className="flex gap-2">
              <ExclamationTriangleIcon /> Important!
            </div>
          </AlertTitle>
          <AlertDescription>
            Password can't be reset or recovered. Make sure to back it up!
          </AlertDescription>
        </Alert>
        <form
          onSubmit={onSubmit}
          className="w-full md:w-[373px] flex flex-col gap-8"
        >
          <div className="grid gap-1.5">
            <Label htmlFor="current-password">Current Password</Label>
            <Input
              id="current-password"
              type={currentUnlockPasswordVisible ? "text" : "password"}
              name="password"
              onChange={(e) => setCurrentUnlockPassword(e.target.value)}
              value={currentUnlockPassword}
              placeholder="Password"
              endAdornment={
                <PasswordViewAdornment
                  isRevealed={currentUnlockPasswordVisible}
                  onChange={(passwordView) =>
                    setCurrentUnlockPasswordVisible(passwordView)
                  }
                />
              }
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="new-password">New Password</Label>
            <Input
              id="new-password"
              type={newUnlockPasswordVisible ? "text" : "password"}
              name="password"
              onChange={(e) => setNewUnlockPassword(e.target.value)}
              value={newUnlockPassword}
              placeholder="Password"
              endAdornment={
                <PasswordViewAdornment
                  isRevealed={newUnlockPasswordVisible}
                  onChange={(passwordView) =>
                    setNewUnlockPasswordVisible(passwordView)
                  }
                />
              }
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="confirm-new-password">Confirm New Password</Label>
            <Input
              id="confirm-new-password"
              type={confirmNewUnlockPasswordVisible ? "text" : "password"}
              name="password"
              onChange={(e) => setConfirmNewUnlockPassword(e.target.value)}
              value={confirmNewUnlockPassword}
              placeholder="Password"
              endAdornment={
                <PasswordViewAdornment
                  isRevealed={confirmNewUnlockPasswordVisible}
                  onChange={(passwordView) =>
                    setConfirmNewUnlockPasswordVisible(passwordView)
                  }
                />
              }
            />
          </div>
          <div className="flex justify-start">
            <LoadingButton loading={loading}>Change Password</LoadingButton>
          </div>
        </form>
      </div>
    </>
  );
}
