import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";
import ConnectButton from "src/components/ConnectButton";
import { handleRequestError } from "src/utils/handleRequestError";
import { useInfo } from "src/hooks/useInfo";
import Container from "src/components/Container";
import Input from "src/components/Input";

export default function Start() {
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const navigate = useNavigate();
  const { data: csrf } = useCSRF();
  const { mutate: refetchInfo } = useInfo();

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
      handleRequestError("Failed to connect", error);
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <Container>
        <p className="font-light text-center text-md leading-relaxed dark:text-neutral-400 mb-14">
          Use your password to unlock and start NWC
        </p>
        <form onSubmit={onSubmit} className="w-full mb-10">
          <>
            <Input
              name="unlock"
              onChange={(e) => setUnlockPassword(e.target.value)}
              value={unlockPassword}
              type="password"
              placeholder="Password"
            />
            <ConnectButton isConnecting={loading} />
          </>
        </form>
        <Link to="/setup" className=" text-red-500">
          Forgot password?
        </Link>
      </Container>
    </>
  );
}
