import { InfoCircledIcon } from "@radix-ui/react-icons";
import { AlertTriangleIcon } from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

import Container from "src/components/Container";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function BackupNode() {
  const navigate = useNavigate();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();

  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [showPasswordScreen, setShowPasswordScreen] = useState<boolean>(false);
  const [loading, setLoading] = React.useState(false);

  const onSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!csrf) {
      throw new Error("No CSRF token");
    }

    const isHttpMode = window.location.protocol.startsWith("http");

    try {
      setLoading(true);

      if (isHttpMode) {
        const response = await fetch("/api/backup", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            UnlockPassword: unlockPassword,
          }),
        });

        if (!response?.ok) {
          throw new Error(`Error:${response?.statusText}`);
        }
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = "nwc.bkp";
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        a.remove();
      } else {
        await request("/api/backup", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
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
        title="Migrate Your Node"
        description="Your Alby Hub will be stopped and you will receive a backup file you can import on another host or machine."
      />
      <Alert>
        <AlertTriangleIcon className="h-4 w-4" />
        <AlertTitle>Do not run your node on multiple devices</AlertTitle>
        <AlertDescription>
          Your node maintains channel state with your channel partners. After
          you create this backup, do not restart Alby Hub on this device.
        </AlertDescription>
      </Alert>
      <Alert>
        <InfoCircledIcon className="h-4 w-4" />
        <AlertTitle>What Happens Next</AlertTitle>
        <AlertDescription>
          You'll need to enter your unlock password to download and decrypt a
          backup of your Alby Hub data. After your backup is downloaded, we'll
          give you instructions on how to import the backup file on another host
          or machine.
        </AlertDescription>
      </Alert>
      {showPasswordScreen ? (
        <Container>
          <h1 className="text-xl font-medium">Enter unlock password</h1>
          <p className="text-center text-md text-muted-foreground mb-14">
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
        <div className="flex justify-center">
          <Button
            type="submit"
            disabled={loading}
            size="lg"
            className="w-full"
            onClick={() => setShowPasswordScreen(true)}
          >
            Create Backup To Migrate Node
          </Button>
        </div>
      )}
    </>
  );
}
