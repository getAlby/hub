import { AlertTriangle } from "lucide-react";
import React from "react";

import Container from "src/components/Container";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

export function AutoUnlock() {
  const { toast } = useToast();
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
      toast({
        title: `Successfully ${unlockPassword ? "enabled" : "disabled"} auto-unlock`,
      });
    } catch (error) {
      toast({
        title: "Auto Unlock change failed",
        description: (error as Error).message,
        variant: "destructive",
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
      <Container>
        {!info.autoUnlockPasswordEnabled && (
          <>
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Security Warning</AlertTitle>
              <AlertDescription>
                By enabling this feature your unlock password will be saved in
                plaintext on the device running this hub. If the device is
                compromised,{" "}
                <span className="underline">your funds will be lost</span>.{" "}
                <span className="font-semibold">
                  Do not enable this feature if you do not physically own the
                  machine Alby Hub is running on.
                </span>
              </AlertDescription>
            </Alert>
            <form
              onSubmit={onSubmit}
              className="w-full flex flex-col gap-3 mt-3"
            >
              <div className="grid gap-1.5">
                <Label htmlFor="unlock-password">Unlock Password</Label>
                <Input
                  id="unlock-password"
                  type="password"
                  name="password"
                  onChange={(e) => setUnlockPassword(e.target.value)}
                  value={unlockPassword}
                  placeholder="Password"
                />
              </div>

              <LoadingButton loading={loading}>
                Enable Auto Unlock
              </LoadingButton>
            </form>
          </>
        )}
        {info.autoUnlockPasswordEnabled && (
          <>
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Security Warning</AlertTitle>
              <AlertDescription>
                Your unlock password is currently saved in plaintext on the
                device running this hub. If your device is compromised, your
                funds will be lost.
              </AlertDescription>
            </Alert>
            <form
              onSubmit={onSubmit}
              className="w-full flex flex-col gap-3 mt-3"
            >
              <LoadingButton loading={loading}>
                Disable Auto Unlock
              </LoadingButton>
            </form>
          </>
        )}
      </Container>
    </>
  );
}
