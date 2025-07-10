import { RefreshCwIcon } from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import PasswordInput from "src/components/password/PasswordInput";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { UnlinkAlbyAccount } from "src/components/UnlinkAlbyAccount";

import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request"; // build the project for this to appear

type AuthCodeFormProps = {
  url: string;
};

function AuthCodeForm({ url }: AuthCodeFormProps) {
  const [authCode, setAuthCode] = useState("");
  const navigate = useNavigate();
  const { toast } = useToast();

  const { data: info, mutate: refetchInfo } = useInfo();

  const [hasRequestedCode, setRequestedCode] = React.useState(false);
  const [isLoading, setLoading] = React.useState(false);

  async function requestAuthCode() {
    setRequestedCode((hasRequestedCode) => {
      if (!url) {
        return false;
      }
      if (!hasRequestedCode) {
        openLink(url);
      }
      return true;
    });
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      await request(`/api/alby/callback?code=${authCode}`, {
        headers: {
          "Content-Type": "application/json",
        },
      });
      await refetchInfo();
      navigate("/");
    } catch (error) {
      handleRequestError(toast, "Failed to connect", error);
    }
    setLoading(false);
  }

  return (
    <Container>
      <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
        <div className="grid gap-5">
          <TwoColumnLayoutHeader
            title="Connect your Alby Account"
            description="A new window will open. Sign in with your Alby Account, copy the Authorization Code, and paste it here."
          />
          {!hasRequestedCode && (
            <>
              <Button onClick={requestAuthCode}>
                Request Authorization Code
              </Button>
            </>
          )}
          {hasRequestedCode && (
            <>
              <div className="grid gap-4 w-full">
                <div className="grid gap-1.5">
                  <Label htmlFor="authorization-code">Authorization Code</Label>
                  <PasswordInput
                    autoFocus
                    id="authorization-code"
                    placeholder="Enter code you see in the browser"
                    onChange={setAuthCode}
                    value={authCode}
                  />
                </div>
              </div>
              <div className="flex gap-4">
                <LoadingButton loading={isLoading} className="flex-1">
                  Submit
                </LoadingButton>
                <Button
                  type="button"
                  size="icon"
                  variant="outline"
                  onClick={() => url && openLink(url)}
                >
                  <RefreshCwIcon className="size-4" />
                </Button>
              </div>
            </>
          )}
          {info?.albyUserIdentifier && (
            <div className="flex flex-col justify-center items-center mt-15 gap-5">
              <p className="text-muted-foreground">or</p>
              <UnlinkAlbyAccount>
                <Button variant="outline" size="sm">
                  Disconnect Alby Account
                </Button>
              </UnlinkAlbyAccount>
            </div>
          )}
        </div>
      </form>
    </Container>
  );
}

export default AuthCodeForm;
