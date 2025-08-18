import React from "react";
import Container from "src/components/Container";
import PasswordInput from "src/components/password/PasswordInput";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";

import { useInfo } from "src/hooks/useInfo";
import { saveAuthToken } from "src/lib/auth";
import { AuthTokenResponse } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export default function Start() {
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [buttonText, setButtonText] = React.useState("Start");

  const { data: info } = useInfo(true); // poll the info endpoint to auto-redirect when app is running

  const { toast } = useToast();

  const startupState = info?.startupState;
  const startupError = info?.startupError;
  const startupErrorTime = info?.startupErrorTime;

  React.useEffect(() => {
    if (startupState) {
      setButtonText(startupState);
    }
  }, [startupState]);

  React.useEffect(() => {
    if (startupError && startupErrorTime) {
      toast({
        title: "Failed to start",
        description: startupError,
        variant: "destructive",
      });
      setLoading(false);
      setButtonText("Start");
      setUnlockPassword("");
    }
  }, [startupError, toast, startupErrorTime]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      setLoading(true);
      setButtonText("Please wait...");

      const authTokenResponse = await request<AuthTokenResponse>("/api/start", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          unlockPassword,
        }),
      });
      if (authTokenResponse) {
        saveAuthToken(authTokenResponse.token);
      }
    } catch (error) {
      handleRequestError(toast, "Failed to connect", error);
      setLoading(false);
      setButtonText("Start");
      setUnlockPassword("");
    }
  }

  return (
    <>
      <Container>
        <div className="mx-auto grid gap-5">
          <TwoColumnLayoutHeader
            title="Start"
            description="Enter your password to unlock and start Alby Hub."
          />
          <form onSubmit={onSubmit}>
            <div className="grid gap-4">
              <div className="grid gap-1.5">
                <Label htmlFor="password">Password</Label>
                <PasswordInput
                  id="password"
                  onChange={setUnlockPassword}
                  autoFocus
                  value={unlockPassword}
                />
              </div>
              <LoadingButton
                type="submit"
                loading={loading}
                disabled={!unlockPassword}
              >
                {buttonText}
              </LoadingButton>
              {loading && (
                <p className="text-muted-foreground text-xs text-center">
                  Starting Alby Hub may take a few minutes.
                </p>
              )}
            </div>
          </form>
        </div>
      </Container>
    </>
  );
}
