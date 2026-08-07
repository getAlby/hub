import { InfoIcon, TriangleAlertIcon } from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router";
import PasswordInput from "src/components/password/PasswordInput";

import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";

import { handleRequestError } from "src/utils/handleRequestError";
import { isHttpMode } from "src/utils/isHttpMode";
import { request } from "src/utils/request";

export function MigrateNode() {
  const navigate = useNavigate();

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
      handleRequestError("Failed to backup the node", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <SettingsHeader
        title="Download Hub backup"
        pageTitle="Hub backup"
        description="Move Alby Hub to another device. Your Lightning funds stay on the node backend (e.g. Greenlight) until you leave that backend."
      />

      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <TriangleAlertIcon className="size-4" />
            <h3>Do not run your Alby Hub on multiple devices</h3>
          </div>
          <p className="text-sm ml-7">
            After creating this file, stop using Alby Hub on this device before
            unlocking on the new one.
          </p>
        </div>
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <TriangleAlertIcon className="size-4" />
            <h3>Restore only on a fresh Alby Hub</h3>
          </div>
          <p className="text-sm ml-7">
            On the new device choose &quot;I already have a Hub&quot; then
            &quot;Restore Hub backup&quot;.
          </p>
        </div>
        <div className="flex flex-col gap-1">
          <div className="flex gap-3 items-center">
            <InfoIcon className="size-4" />
            <h3>What is included?</h3>
          </div>
          <p className="text-sm ml-7">
            Encrypted Hub data: apps, settings, and keys. On Greenlight, channels
            stay in the cloud; this file reconnects you to the same node.
          </p>
        </div>
      </div>

      {showPasswordScreen ? (
        <div>
          <h1 className="font-medium mb-1">Enter unlock password</h1>
          <p className="text-muted-foreground mb-4">
            Your unlock password encrypts the backup file
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
            Download Hub backup
          </Button>

          <LinkButton to="/settings/backup" variant={"secondary"}>
            Back to Backup
          </LinkButton>
        </div>
      )}
    </>
  );
}
