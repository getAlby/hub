import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { toast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request"; // build the project for this to appear

function AuthCodeForm() {
  const [authCode, setAuthCode] = useState("");
  const navigate = useNavigate();
  const { data: csrf } = useCSRF();
  const { data: info } = useInfo();
  const { mutate: refetchInfo } = useInfo();

  const [hasRequestedCode, setRequestedCode] = React.useState(false);
  const [isLoading, setLoading] = React.useState(false);

  async function requestAuthCode() {
    setRequestedCode((hasRequestedCode) => {
      if (!info) {
        return false;
      }
      if (!hasRequestedCode) {
        openLink(info.albyAuthUrl);
      }
      return true;
    });
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      if (!csrf) {
        throw new Error("info not loaded");
      }
      await request(`/api/alby/callback?code=${authCode}`, {
        headers: {
          "X-CSRF-Token": csrf,
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
            title="Alby OAuth"
            description="Enter your Auth Code to connect to Alby"
          />
          {!hasRequestedCode && (
            <>
              <Button onClick={requestAuthCode}>Request Auth Code</Button>
            </>
          )}
          {hasRequestedCode && (
            <>
              <div className="grid gap-4 w-full">
                <div className="grid gap-1.5">
                  <Label htmlFor="authorization-code">Authorization Code</Label>
                  <Input
                    type="text"
                    name="authorization-code"
                    id="authorization-code"
                    placeholder="Enter code you see in the browser"
                    value={authCode}
                    onChange={(e) => setAuthCode(e.target.value)}
                    required={true}
                  />
                </div>
              </div>
              <LoadingButton loading={isLoading}>Submit</LoadingButton>
            </>
          )}
        </div>
      </form>
    </Container>
  );
}

export default AuthCodeForm;
