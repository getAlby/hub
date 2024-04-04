import React from "react";

import Container from "src/components/Container";
import SettingsHeader from "src/components/SettingsHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

export function ChangeUnlockPassword() {
  const { data: csrf } = useCSRF();
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
      if (!csrf) {
        throw new Error("No CSRF token");
      }
      if (newUnlockPassword !== confirmNewUnlockPassword) {
        throw new Error("Password confirmation does not match");
      }
      setLoading(true);
      await request("/api/unlock-password", {
        method: "PATCH",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          currentUnlockPassword,
          newUnlockPassword,
        }),
      });
      await refetchInfo();
      toast({
        title: "Password changed!",
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
        title="Change Unlock Password"
        description="Enter your current and new unlock password. Your node
          will be stopped as part of this process."
      />
      <Container>
        <form onSubmit={onSubmit} className="w-full flex flex-col gap-3">
          <div className="grid gap-1.5">
            <Label htmlFor="current-password">Current Password</Label>
            <Input
              id="current-password"
              type="password"
              name="password"
              onChange={(e) => setCurrentUnlockPassword(e.target.value)}
              value={currentUnlockPassword}
              placeholder="Password"
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="new-password">New Password</Label>
            <Input
              id="new-password"
              type="password"
              name="password"
              onChange={(e) => setNewUnlockPassword(e.target.value)}
              value={newUnlockPassword}
              placeholder="Password"
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="confirm-new-password">Confirm New Password</Label>
            <Input
              id="confirm-new-password"
              type="password"
              name="password"
              onChange={(e) => setConfirmNewUnlockPassword(e.target.value)}
              value={confirmNewUnlockPassword}
              placeholder="Password"
            />
          </div>
          <LoadingButton loading={loading}>Change Password</LoadingButton>
        </form>
      </Container>
    </>
  );
}
