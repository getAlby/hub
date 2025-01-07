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
        <p className="text-muted-foreground">
          In some situations it can be impractical to manually unlock the wallet
          every time Alby Hub is started. In those cases you can save the unlock
          password in plaintext so that Alby Hub can auto-unlock itself.
        </p>
        <Alert className="mt-3">
          <AlertTriangle className="h-4 w-4" />
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
