import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

export default function Unlock() {
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { data: csrf } = useCSRF();
  const { data: info } = useInfo();
  const { mutate: refetchInfo } = useInfo();
  const { toast } = useToast();

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
      toast({
        title: (error as Error).message,
        variant: "destructive",
      });
      // handleRequestError("Failed to connect", error);
    } finally {
      setLoading(false);
    }
  }

  if (loading) return <Loading />;

  return (
    <>
      <form onSubmit={onSubmit} className="w-full">
        <div className="mx-auto grid w-[350px] gap-6">
          <div className="grid gap-2 text-center">
            <h1 className="text-3xl font-bold">Login</h1>
            <p className="text-balance text-muted-foreground">
              Enter your unlock password to get started
            </p>
          </div>
          <div className="grid gap-4">
            <div className="grid gap-2">
              <div className="flex items-center">
                <Label htmlFor="password">Password</Label>
              </div>
              <Input
                id="password"
                type="password"
                required
                value={unlockPassword}
                onChange={(e) => setUnlockPassword(e.target.value)}
              />
            </div>
            <Button type="submit" className="w-full">
              Login
            </Button>
          </div>
        </div>
      </form>
    </>
  );
}
