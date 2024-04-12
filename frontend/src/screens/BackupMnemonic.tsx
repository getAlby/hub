import { LifeBuoy, ShieldAlert, ShieldCheck } from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

import Container from "src/components/Container";
import Loading from "src/components/Loading";
import MnemonicInputs from "src/components/MnemonicInputs";
import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useEncryptedMnemonic } from "src/hooks/useEncryptedMnemonic";
import { useInfo } from "src/hooks/useInfo";
import { aesGcmDecrypt } from "src/utils/aesgcm";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function BackupMnemonic() {
  const navigate = useNavigate();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const { mutate: refetchInfo } = useInfo();
  const { data: mnemonic } = useEncryptedMnemonic();

  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [decryptedMnemonic, setDecryptedMnemonic] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [backedUp, setIsBackedUp] = useState<boolean>(false);

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
      toast({
        title: "Incorrect password",
        description: "Failed to decrypt mnemonic.",
        variant: "destructive",
      });
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

      toast({ title: "Recovery phrase backed up!" });
    } catch (error) {
      handleRequestError(toast, "Failed to store back up info", error);
    }
  }

  return (
    <>
      <SettingsHeader
        title="Backup Your Keys"
        description="Make sure to store them somewhere safe."
      />
      {!decryptedMnemonic ? (
        <Container>
          <h1 className="text-xl font-medium">Please confirm it's you</h1>
          <p className="text-center text-md text-muted-foreground mb-14">
            Enter your unlock password to continue
          </p>
          <form
            onSubmit={onSubmitPassword}
            className="w-full flex flex-col gap-3"
          >
            <>
              <div className="grid gap-1.5">
                <Label htmlFor="password">Password</Label>
                <Input
                  type="password"
                  name="password"
                  onChange={(e) => setUnlockPassword(e.target.value)}
                  value={unlockPassword}
                  placeholder="Password"
                />
              </div>
              <LoadingButton loading={loading}>Continue</LoadingButton>
            </>
          </form>
        </Container>
      ) : (
        <form
          onSubmit={onSubmit}
          className="flex mt-6 flex-col gap-2 mx-auto max-w-2xl text-sm"
        >
          <div className="flex flex-col gap-4 mb-4 text-muted-foreground">
            <div className="flex gap-2 items-center ">
              <div className="shrink-0 ">
                <LifeBuoy className="w-6 h-6" />
              </div>
              <span>
                Your recovery phrase is a set of 12 words that{" "}
                <b>backs up your wallet</b>
              </span>
            </div>
            <div className="flex gap-2 items-center">
              <div className="shrink-0">
                <ShieldCheck className="w-6 h-6" />
              </div>
              <span>
                Make sure to write them down somewhere safe and private
              </span>
            </div>
            <div className="flex gap-2 items-center text-destructive">
              <div className="shrink-0 ">
                <ShieldAlert className="w-6 h-6" />
              </div>
              <span>
                If you lose your recovery phrase, you will lose access to your
                funds
              </span>
            </div>
          </div>

          <MnemonicInputs mnemonic={decryptedMnemonic} readOnly={true}>
            <div className="flex items-center mt-5">
              <Checkbox
                id="backup"
                onCheckedChange={() => setIsBackedUp(!backedUp)}
              />
              <Label htmlFor="backup" className="ml-2">
                I've backed my recovery phrase to my wallet in a private and
                secure place
              </Label>
            </div>
          </MnemonicInputs>
          <div className="flex justify-center">
            <Button type="submit" disabled={!backedUp} size="lg">
              Continue
            </Button>
          </div>
        </form>
      )}
    </>
  );
}
