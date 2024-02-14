import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";
import ConnectButton from "src/components/ConnectButton";
import { handleRequestError } from "src/utils/handleRequestError";
import { useInfo } from "src/hooks/useInfo";
import Container from "src/components/Container";

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
      const res = await request("/api/unlock", {
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
        <p className="font-light text-center text-md leading-relaxed dark:text-neutral-400 mb-14">
          Use your password to unlock NWC
        </p>
        <form onSubmit={onSubmit} className="w-full">
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
      </Container>
    </>
  );
}
