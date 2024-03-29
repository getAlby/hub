import {
  PopiconsLifebuoyLine,
  PopiconsShieldLine,
  PopiconsTriangleExclamationLine,
} from "@popicons/react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Input from "src/components/Input";

import Container from "src/components/Container";
import Loading from "src/components/Loading";
import MnemonicInputs from "src/components/MnemonicInputs";
import PasswordViewAdornment from "src/components/PasswordAdornment";
import toast from "src/components/Toast";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/loading-button";
import { useCSRF } from "src/hooks/useCSRF";
import { useEncryptedMnemonic } from "src/hooks/useEncryptedMnemonic";
import { useInfo } from "src/hooks/useInfo";
import { aesGcmDecrypt } from "src/utils/aesgcm";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function BackupMnemonic() {
  const navigate = useNavigate();
  const { data: csrf } = useCSRF();
  const { mutate: refetchInfo } = useInfo();
  const { data: mnemonic } = useEncryptedMnemonic();

  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [passwordVisible, setPasswordVisible] = React.useState(false);
  const [decryptedMnemonic, setDecryptedMnemonic] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [backedUp, isBackedUp] = useState<boolean>(false);

  if (!mnemonic) {
    return <Loading />;
  }

  const onSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      const dec = await aesGcmDecrypt(mnemonic.mnemonic, unlockPassword);
      setDecryptedMnemonic(dec);
    } catch (error) {
      toast.error("Failed to decrypt mnemonic: incorrect password");
    } finally {
      setLoading(false);
    }
  };

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!csrf) {
      throw new Error("No CSRF token");
    }

    const currentDate = new Date();
    const sixMonthsLater = new Date(
      currentDate.setMonth(currentDate.getMonth() + 6)
    );

    try {
      await request("/api/backup-reminder", {
        method: "PATCH",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          nextBackupReminder: sixMonthsLater.toISOString(),
        }),
      });
      await refetchInfo();

      navigate("/");
      toast.success("Recovery phrase backed up!");
    } catch (error) {
      handleRequestError("Failed to store back up info", error);
    }
  }

  return (
    <>
      {!decryptedMnemonic ? (
        <Container>
          <p className="font-light text-center text-md leading-relaxed dark:text-neutral-400 mb-14">
            Enter your unlock password to continue
          </p>
          <form
            onSubmit={onSubmitPassword}
            className="w-full flex flex-col gap-3"
          >
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
              <LoadingButton loading={loading}>Unlock</LoadingButton>
            </>
          </form>
        </Container>
      ) : (
        <form
          onSubmit={onSubmit}
          className="flex mt-6 flex-col gap-2 mx-auto max-w-2xl text-sm"
        >
          <h1 className="font-semibold text-2xl font-headline mb-">
            Back up your wallet
          </h1>

          <div className="flex flex-col gap-4 mb-4 text-muted-foreground">
            <div className="flex gap-2 items-center ">
              <div className="shrink-0 ">
                <PopiconsLifebuoyLine className="w-6 h-6" />
              </div>
              <span>
                Your recovery phrase is a set of 12 words that{" "}
                <b>backs up your wallet</b>
              </span>
            </div>
            <div className="flex gap-2 items-center">
              <div className="shrink-0">
                <PopiconsShieldLine className="w-6 h-6" />
              </div>
              <span>
                Make sure to write them down somewhere safe and private
              </span>
            </div>
            <div className="flex gap-2 items-center text-destructive">
              <div className="shrink-0 ">
                <PopiconsTriangleExclamationLine className="w-6 h-6" />
              </div>
              <span>
                If you lose your recovery phrase, you will lose access to your
                funds
              </span>
            </div>
          </div>

          <MnemonicInputs mnemonic={decryptedMnemonic} readOnly={true}>
            <div className="flex items-center mt-5">
              <input
                id="checkbox"
                type="checkbox"
                onChange={(event) => {
                  isBackedUp(event.target.checked);
                }}
                checked={backedUp}
                className="w-4 h-4 text-purple-700 bg-gray-100 border-gray-300 rounded focus:ring-purple-700 dark:focus:ring-purple-800 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
              />
              <label
                htmlFor="checkbox"
                className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-100"
              >
                I've backed my recovery phrase to my wallet in a private and
                secure place
              </label>
            </div>
          </MnemonicInputs>
          <div className="flex justify-center">
            <Button disabled={!backedUp} size="lg">
              Continue
            </Button>
          </div>
        </form>
      )}
    </>
  );
}
