import { Eye, EyeOff } from "lucide-react";
import React from "react";

import Container from "src/components/Container";
import SettingsHeader from "src/components/SettingsHeader";
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

  const [showCurrentPassword, setShowCurrentPassword] =
    React.useState<boolean>(false);
  const [showNewPassword, setShowNewPassword] = React.useState<boolean>(false);
  const [showConfirmPassword, setShowConfirmPassword] =
    React.useState<boolean>(false);

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
        title="Change Unlock Password"
        description="Enter your current and new unlock password. Your node
          will be stopped as part of this process."
      />
      <Container>
        <form onSubmit={onSubmit} className="w-full flex flex-col gap-3">
          <div className="grid gap-1.5">
            <Label htmlFor="current-password">Current Password</Label>
            <div className="relative">
              <Input
                id="current-password"
                type={showCurrentPassword ? "text" : "password"}
                name="password"
                onChange={(e) => setCurrentUnlockPassword(e.target.value)}
                value={currentUnlockPassword}
                placeholder="Password"
              />
              <button
                type="button"
                className="absolute right-2.5 top-2.5"
                onClick={() => setShowCurrentPassword(!showCurrentPassword)}
              >
                {showCurrentPassword ? (
                  <EyeOff className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <Eye className="h-4 w-4 text-muted-foreground" />
                )}
              </button>
            </div>
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="new-password">New Password</Label>
            <div className="relative">
              <Input
                id="new-password"
                type={showNewPassword ? "text" : "password"}
                name="password"
                onChange={(e) => setNewUnlockPassword(e.target.value)}
                value={newUnlockPassword}
                placeholder="Password"
              />
              <button
                type="button"
                className="absolute right-2.5 top-2.5"
                onClick={() => setShowNewPassword(!showNewPassword)}
              >
                {showNewPassword ? (
                  <EyeOff className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <Eye className="h-4 w-4 text-muted-foreground" />
                )}
              </button>
            </div>
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="confirm-new-password">Confirm New Password</Label>
            <div className="relative">
              <Input
                id="confirm-new-password"
                type={showConfirmPassword ? "text" : "password"}
                name="password"
                onChange={(e) => setConfirmNewUnlockPassword(e.target.value)}
                value={confirmNewUnlockPassword}
                placeholder="Password"
              />
              <button
                type="button"
                className="absolute right-2.5 top-2.5"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              >
                {showConfirmPassword ? (
                  <EyeOff className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <Eye className="h-4 w-4 text-muted-foreground" />
                )}
              </button>
            </div>
          </div>
          <LoadingButton loading={loading}>Change Password</LoadingButton>
        </form>
      </Container>
    </>
  );
}
