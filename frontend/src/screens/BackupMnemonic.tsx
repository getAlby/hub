import {
  CopyIcon,
  ExternalLinkIcon,
  LifeBuoy,
  ShieldAlert,
} from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

import ExternalLink from "src/components/ExternalLink";
import MnemonicInputs from "src/components/MnemonicInputs";
import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { copyToClipboard } from "src/lib/clipboard";
import { MnemonicResponse } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function BackupMnemonic() {
  const navigate = useNavigate();

  const { toast } = useToast();
  const { mutate: refetchInfo } = useInfo();

  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [decryptedMnemonic, setDecryptedMnemonic] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [backedUp, setIsBackedUp] = useState<boolean>(false);
  const [backedUp2, setIsBackedUp2] = useState<boolean>(false);
  const { data: info } = useInfo();

  const onSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      const result = await request<MnemonicResponse>("/api/mnemonic", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          unlockPassword,
        }),
      });

      setDecryptedMnemonic(result?.mnemonic ?? "");
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

    const currentDate = new Date();
    const sixMonthsLater = new Date(
      currentDate.setMonth(currentDate.getMonth() + 6)
    );

    try {
      await request("/api/backup-reminder", {
        method: "PATCH",
        headers: {
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
        title="Key Backup"
        description="Key recovery phrase is a group of 12 random words that are the only way to recover access to your wallet on another machine or when you loose your unlock password."
      />
      {!decryptedMnemonic ? (
        <div>
          <p className="text-muted-foreground">
            Enter your unlock password to view your recovery phrase.
          </p>
          <form
            onSubmit={onSubmitPassword}
            className="max-w-md flex flex-col gap-3 mt-8"
          >
            <div className="grid gap-1.5 mb-8">
              <Label htmlFor="password">Password</Label>
              <Input
                type="password"
                name="password"
                onChange={(e) => setUnlockPassword(e.target.value)}
                value={unlockPassword}
                placeholder="Password"
              />
            </div>
            <div className="flex justify-start">
              <LoadingButton loading={loading}>
                View Recovery Phrase
              </LoadingButton>
            </div>
          </form>
        </div>
      ) : (
        <form
          onSubmit={onSubmit}
          className="flex mt-6 flex-col gap-2 mx-auto max-w-md text-sm"
        >
          <div className="flex flex-col gap-4 mb-4 text-muted-foreground">
            <div className="flex gap-2 items-center ">
              <div className="shrink-0 ">
                <LifeBuoy className="w-6 h-6" />
              </div>
              <span>
                Your recovery phrase is a set of 12 words that{" "}
                <b>backs up your wallet on-chain balance</b>.&nbsp;
                {info?.albyAccountConnected && (
                  <>
                    Channel backups are saved automatically to your Alby
                    Account, encrypted with your recovery phrase.
                  </>
                )}
                {!info?.albyAccountConnected && (
                  <>
                    Make sure to also backup your <b>data directory</b> as this
                    is required to recover funds on your channels. You can also
                    connect your Alby Account for automatic encrypted backups.
                  </>
                )}
              </span>
            </div>
            <div className="flex gap-2 items-center text-destructive">
              <div className="shrink-0 ">
                <ShieldAlert className="w-6 h-6" />
              </div>
              <span>
                If you lose access to your hub and do not have your{" "}
                <b>recovery phrase</b>
                {!info?.albyAccountConnected && (
                  <>&nbsp;or do not backup your data directory</>
                )}
                , you will lose access to your funds.
              </span>
            </div>
          </div>
          <div className="mb-5">
            <ExternalLink
              className="underline flex items-center"
              to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/backups"
            >
              Learn more about backups
              <ExternalLinkIcon className="w-4 h-4 ml-2" />
            </ExternalLink>
          </div>

          <MnemonicInputs
            mnemonic={decryptedMnemonic}
            readOnly={true}
            description="  Writing these words down, store them somewhere safe and keep them
            secret."
          >
            <div className="flex items-center mt-5 text-muted-foreground">
              <Checkbox
                id="backup"
                required
                onCheckedChange={() => setIsBackedUp(!backedUp)}
              />
              <Label htmlFor="backup" className="ml-2">
                I've backed up my recovery phrase to my wallet in a private and
                secure place
              </Label>
            </div>
            {backedUp && !info?.albyAccountConnected && (
              <div className="flex mt-5 text-muted-foreground">
                <Checkbox
                  id="backup2"
                  required
                  onCheckedChange={() => setIsBackedUp2(!backedUp2)}
                />
                <Label htmlFor="backup2" className="ml-2">
                  I understand the <b>recovery phrase</b> AND{" "}
                  <b>a backup of my hub data directory</b> is required to
                  recover funds from my lightning channels.{" "}
                </Label>
              </div>
            )}
            <Button
              type="button"
              variant={"destructive"}
              className="flex gap-2 items-center justify-center mt-5 mx-auto"
              onClick={() => copyToClipboard(decryptedMnemonic, toast)}
            >
              <CopyIcon className="w-4 h-4 mr-2" />
              Dangerously Copy
            </Button>
          </MnemonicInputs>
          <div className="flex justify-center">
            <Button type="submit" size="lg">
              Continue
            </Button>
          </div>
        </form>
      )}
    </>
  );
}
