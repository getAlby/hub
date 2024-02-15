import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";
import ConnectButton from "src/components/ConnectButton";
import { handleRequestError } from "src/utils/handleRequestError";
import { useInfo } from "src/hooks/useInfo";
import Container from "src/components/Container";
import Input from "src/components/Input";
import PasswordViewAdornment from "src/components/PasswordAdornment";

export default function Unlock() {
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [passwordVisible, setPasswordVisible] = React.useState(false);
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
            <Input
              name="unlock"
              onChange={(e) => setUnlockPassword(e.target.value)}
              value={unlockPassword}
              type={passwordVisible ? "text" : "password"}
              placeholder="Password"
              endAdornment={
                <PasswordViewAdornment
                  onChange={(passwordView) => {
                    setPasswordVisible(passwordView);
                  }}
                />
              }
            />
            <ConnectButton isConnecting={loading} />
          </>
        </form>
      </Container>
    </>
  );
}
