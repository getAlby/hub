import React from "react";
import { useNavigate } from "react-router-dom";
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

export default function Start() {
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const navigate = useNavigate();
  const { data: csrf } = useCSRF();
  const { mutate: refetchInfo } = useInfo();
  const { toast } = useToast();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      setLoading(true);
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
      await refetchInfo();

      navigate("/");
    } catch (error) {
      handleRequestError(toast, "Failed to connect", error);
    } finally {
      setLoading(false);
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
                Login
              </LoadingButton>
            </div>
          </form>
        </div>
      </Container>
    </>
  );
}
