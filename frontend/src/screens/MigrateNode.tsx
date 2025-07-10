import { InfoIcon, TriangleAlertIcon } from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import PasswordInput from "src/components/password/PasswordInput";

import SettingsHeader from "src/components/SettingsHeader";
import { Button, LinkButton } from "src/components/ui/button";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { handleRequestError } from "src/utils/handleRequestError";
import { isHttpMode } from "src/utils/isHttpMode";
import { request } from "src/utils/request";

export function MigrateNode() {
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

      navigate("/create-node-migration-file-success");
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
        description="Create migration file in order to move your Alby Hub to another device or server."
      />

      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <TriangleAlertIcon className="size-4" />
            <h3>Do not run your Alby Hub on multiple devices</h3>
          </div>
          <p className="text-sm ml-7">
            After creating this migration file, do not restart Alby Hub on this
            device, as this will cause problems and may cause force channel
            closures.
          </p>
        </div>
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <TriangleAlertIcon className="size-4" />
            <h3>Migrate this file only to fresh Alby Hub</h3>
          </div>
          <p className="text-sm ml-7">
            To import the migration file, you must have a brand new Alby Hub on
            another device and use the “Advanced” option during the onboarding.
          </p>
        </div>
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <InfoIcon className="size-4" />
            <h3>What happens next?</h3>
          </div>
          <p className="text-sm ml-7">
            After typing your unlock password, you'll be able to download a
            migration file which contains a backup of your Alby Hub data. Then
            you'll see instructions on how to import this migration file into
            another device or server.
          </p>
        </div>
      </div>

      {showPasswordScreen ? (
        <div>
          <h1 className="font-medium mb-1">Enter unlock password</h1>
          <p className="text-muted-foreground mb-4">
            Your unlock password will be used to encrypt your migration file
          </p>
          <form
            onSubmit={onSubmitPassword}
            className="w-full md:w-96 flex flex-col gap-6"
          >
            <>
              <div className="grid gap-2">
                <Label htmlFor="password">Password</Label>
                <PasswordInput
                  id="password"
                  autoFocus
                  onChange={setUnlockPassword}
                  value={unlockPassword}
                />
              </div>
            </>
            <LoadingButton loading={loading}>Continue</LoadingButton>
          </form>
        </div>
      ) : (
        <div className="flex gap-4">
          <Button
            type="submit"
            disabled={loading}
            onClick={() => setShowPasswordScreen(true)}
          >
            Create Alby Hub Migration File
          </Button>

          <LinkButton to="/settings/backup" variant={"secondary"}>
            Backup Without Migrating Alby Hub
          </LinkButton>
        </div>
      )}
    </>
  );
}
