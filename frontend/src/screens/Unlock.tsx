import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { toast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export default function Unlock() {
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { data: csrf } = useCSRF();
  const { data: info } = useInfo();
  const { mutate: refetchInfo } = useInfo();

  React.useEffect(() => {
    if (!info || info.running) {
      return;
    }
    navigate("/");
  }, [info, location, navigate]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      setLoading(true);
      if (!csrf) {
        throw new Error("info not loaded");
      }
      await request("/api/unlock", {
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
      <form onSubmit={onSubmit} className="w-full p-5">
        <div className="mx-auto grid w-80 max-w-full gap-6">
          <TwoColumnLayoutHeader
            title="Login"
            description=" Enter your unlock password to continue"
          />
          <div className="grid gap-4">
            <div className="grid gap-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                required
                value={unlockPassword}
                onChange={(e) => setUnlockPassword(e.target.value)}
              />
            </div>
            <LoadingButton type="submit" loading={loading}>
              Login
            </LoadingButton>
          </div>
        </div>
      </form>
    </>
  );
}
