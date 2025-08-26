import { AlertTriangleIcon } from "lucide-react";
import React from "react";

import { toast } from "sonner";
import Loading from "src/components/Loading";
import PasswordInput from "src/components/password/PasswordInput";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";

import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

export function AutoUnlock() {
  const { data: info, mutate: refetchInfo } = useInfo();

  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);
      await request("/api/auto-unlock", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          unlockPassword,
        }),
      });
      await refetchInfo();
      setUnlockPassword("");
      toast(
        `Successfully ${unlockPassword ? "enabled" : "disabled"} auto-unlock`
      );
    } catch (error) {
      toast("Auto Unlock change failed", {
        description: (error as Error).message,
      });
    } finally {
      setLoading(false);
    }
  };

  if (!info) {
    return <Loading />;
  }
  if (!info.autoUnlockPasswordSupported) {
    return <p>Your Hub does not support this feature.</p>;
  }

  return (
    <>
      <SettingsHeader
        title="Auto Unlock"
        description="Configure Alby Hub will automatically unlock on start (e.g. after machine reboot)"
      />
      <div>
        <p className="text-muted-foreground">
          In some situations it can be impractical to manually unlock the wallet
          every time Alby Hub is started. In those cases you can save the unlock
          password in plaintext so that Alby Hub can auto-unlock itself.
        </p>
        <Alert className="mt-3">
          <AlertTriangleIcon />
          <AlertTitle>Attention</AlertTitle>
          <AlertDescription>
            Everyone who has access to the machine running this hub could read
            that password and take your funds. Use this only in a secure
            environment.
          </AlertDescription>
        </Alert>
        {!info.autoUnlockPasswordEnabled && (
          <>
            <form
              onSubmit={onSubmit}
              className="w-full md:w-96 flex flex-col gap-4 mt-4"
            >
              <div className="grid gap-2">
                <Label htmlFor="unlock-password">Unlock Password</Label>
                <PasswordInput
                  id="unlock-password"
                  autoFocus
                  onChange={setUnlockPassword}
                  value={unlockPassword}
                />
              </div>
              <div>
                <LoadingButton loading={loading}>
                  Enable Auto Unlock
                </LoadingButton>
              </div>
            </form>
          </>
        )}
        {info.autoUnlockPasswordEnabled && (
          <>
            <form
              onSubmit={onSubmit}
              className="w-full md:w-96 flex flex-col gap-4 mt-4"
            >
              <div>
                <LoadingButton loading={loading}>
                  Disable Auto Unlock
                </LoadingButton>
              </div>
            </form>
          </>
        )}
      </div>
    </>
  );
}
