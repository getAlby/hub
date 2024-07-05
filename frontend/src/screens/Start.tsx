import React from "react";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

const messages: string[] = [
  "Unlocking",
  "Starting the wallet",
  "Connecting to the network",
  "Syncing",
  "Still syncing, please wait...",
];

export default function Start() {
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [buttonText, setButtonText] = React.useState("Login");
  useInfo(true); // poll the info endpoint to auto-redirect when app is running
  const { data: csrf } = useCSRF();
  const { toast } = useToast();

  React.useEffect(() => {
    if (!loading) {
      return;
    }
    let messageIndex = 1;
    const intervalId = setInterval(() => {
      if (messageIndex < messages.length) {
        setButtonText(messages[messageIndex]);
        messageIndex++;
      } else {
        clearInterval(intervalId);
      }
    }, 5000);

    const timeoutId = setTimeout(() => {
      // if redirection didn't happen in 3 minutes info.running is false
      toast({
        title: "Failed to start",
        description: "Please try starting the node again.",
        variant: "destructive",
      });

      setLoading(false);
      setButtonText("Login");
      setUnlockPassword("");
      return;
    }, 180000); // wait for 3 minutes

    return () => {
      clearInterval(intervalId);
      clearTimeout(timeoutId);
    };
  }, [loading, toast]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      setLoading(true);
      setButtonText(messages[0]);
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      await request("/api/start", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          unlockPassword,
        }),
      });
    } catch (error) {
      handleRequestError(toast, "Failed to connect", error);
      setLoading(false);
      setButtonText("Login");
      setUnlockPassword("");
    }
  }

  return (
    <>
      <Container>
        <div className="mx-auto grid gap-5">
          <TwoColumnLayoutHeader
            title="Login"
            description="Enter your password to unlock and start Alby Hub."
          />
          <form onSubmit={onSubmit}>
            <div className="grid gap-4">
              <div className="grid gap-1.5">
                <Label htmlFor="password">Password</Label>
                <Input
                  name="unlock"
                  autoFocus
                  onChange={(e) => setUnlockPassword(e.target.value)}
                  value={unlockPassword}
                  type="password"
                  placeholder="Password"
                />
              </div>
              <LoadingButton type="submit" loading={loading}>
                {buttonText}
              </LoadingButton>
            </div>
          </form>
        </div>
      </Container>
    </>
  );
}
