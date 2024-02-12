import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";
import ConnectButton from "src/components/ConnectButton";
import { handleRequestError } from "src/utils/handleRequestError";
import { useInfo } from "src/hooks/useInfo";
import Container from "src/components/Container";

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
        throw new Error("info not loaded");
      }
      const res = await request("/api/start", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          unlockPassword,
        }),
      });
      console.log({ res });
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
        <p className="text-center text-md leading-relaxed dark:text-neutral-400 mb-14">
          Use your password to unlock NWC
        </p>
        <form onSubmit={onSubmit} className="w-full mb-10">
          <>
            <input
              name="unlock"
              onChange={(e) => setUnlockPassword(e.target.value)}
              value={unlockPassword}
              type="password"
              placeholder="Password"
              className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
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
