import { InfoCircledIcon } from "@radix-ui/react-icons";
import { TriangleAlertIcon } from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

import Container from "src/components/Container";
import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { handleRequestError } from "src/utils/handleRequestError";
import { isHttpMode } from "src/utils/isHttpMode";
import { request } from "src/utils/request";

export function BackupNode() {
  const navigate = useNavigate();

  const { toast } = useToast();

  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [showPasswordScreen, setShowPasswordScreen] = useState<boolean>(false);
  const [loading, setLoading] = React.useState(false);

  const onSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault();

    const _isHttpMode = isHttpMode();

    try {
      setLoading(true);

      if (_isHttpMode) {
        const response = await fetch("/api/backup", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            UnlockPassword: unlockPassword,
          }),
        });

        if (!response.ok) {
          throw new Error(await response.text());
        }
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = "albyhub.bkp";
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        a.remove();
      } else {
        await request("/api/backup", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            unlockPassword,
          }),
        });
      }

      navigate("/node-backup-success");
    } catch (error) {
      handleRequestError(toast, "Failed to backup the node", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <SettingsHeader
        title="Migrate Alby Hub"
        description="Create backup file in order to migrate your Alby Hub onto another device or server."
      />

      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <TriangleAlertIcon className="w-4 h-4" />
            <h3>Do not run your Alby Hub on multiple devices</h3>
          </div>
          <p className="text-sm ml-7">
            After creating this backup file, do not restart Alby Hub on this
            device, as this will cause problems and may cause force channel
            closures.
          </p>
        </div>
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <TriangleAlertIcon className="w-4 h-4" />
            <h3>Migrate this file only to fresh Alby Hub</h3>
          </div>
          <p className="text-sm ml-7">
            To import the migration file, you must have a brand new Alby Hub on
            another device and use the “Advanced” option during the onboarding.
          </p>
        </div>
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <InfoCircledIcon className="w-4 h-4" />
            <h3>What happens next?</h3>
          </div>
          <p className="text-sm ml-7">
            After typing your unlock password, you’ll be able to to download a
            backup of your Alby Hub data. Then you’ll see instructions on how to
            import the backup file into another device or server.
          </p>
        </div>
      </div>

      {showPasswordScreen ? (
        <Container>
          <h1 className="text-xl font-medium mb-1">Enter unlock password</h1>
          <p className="text-center text-md text-muted-foreground mb-4">
            Your unlock password will be used to encrypt your backup
          </p>
          <form
            onSubmit={onSubmitPassword}
            className="w-full flex flex-col gap-3"
          >
            <>
              <div className="grid gap-1.5">
                <Label htmlFor="password">Password</Label>
                <Input
                  type="password"
                  name="password"
                  onChange={(e) => setUnlockPassword(e.target.value)}
                  value={unlockPassword}
                  placeholder="Password"
                />
              </div>
              <LoadingButton loading={loading}>Continue</LoadingButton>
            </>
          </form>
        </Container>
      ) : (
        <div>
          <Button
            type="submit"
            disabled={loading}
            size="lg"
            onClick={() => setShowPasswordScreen(true)}
          >
            Create Backup to Migrate Alby Hub
          </Button>
        </div>
      )}
    </>
  );
}
