import { CopyIcon, TriangleAlertIcon } from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";

import MnemonicInputs from "src/components/MnemonicInputs";
import PasswordViewAdornment from "src/components/PasswordAdornment";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
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
  const [unlockPasswordVisible, setUnlockPasswordVisible] =
    React.useState(false);
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
      {info?.backendType === "CASHU" && <CashuMnemonicWarning />}
      {!decryptedMnemonic ? (
        <div>
          <p className="text-muted-foreground">
            Enter your unlock password to view your recovery phrase.
          </p>
          <form
            onSubmit={onSubmitPassword}
            className="max-w-md flex flex-col gap-3 mt-8"
          >
            <div className="grid gap-2 mb-6">
              <Label htmlFor="password">Password</Label>
              <Input
                type={unlockPasswordVisible ? "text" : "password"}
                name="password"
                onChange={(e) => setUnlockPassword(e.target.value)}
                value={unlockPassword}
                placeholder="Password"
                endAdornment={
                  <PasswordViewAdornment
                    isRevealed={unlockPasswordVisible}
                    onChange={(passwordView) =>
                      setUnlockPasswordVisible(passwordView)
                    }
                  />
                }
              />
            </div>
            <div className="flex justify-start">
              <LoadingButton
                loading={loading}
                variant={"secondary"}
                size={"lg"}
              >
                View Recovery Phase
              </LoadingButton>
            </div>
          </form>
        </div>
      ) : (
        <form
          onSubmit={onSubmit}
          className="flex mt-6 flex-col gap-2 max-w-md text-sm"
        >
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
                <Label
                  htmlFor="backup2"
                  className="ml-2 text-sm text-muted-foreground"
                >
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

// TODO: remove after 2026-01-01
function CashuMnemonicWarning() {
  const [mnemonicMatches, setMnemonicMatches] = React.useState<boolean>();

  React.useEffect(() => {
    (async () => {
      try {
        const result: { matches: boolean } | undefined = await request(
          "/api/command",
          {
            method: "POST",
            body: JSON.stringify({ command: "checkmnemonic" }),
            headers: {
              "Content-Type": "application/json",
            },
          }
        );
        setMnemonicMatches(result?.matches);
      } catch (error) {
        console.error(error);
      }
    })();
  }, []);

  if (mnemonicMatches === undefined) {
    return <Loading />;
  }

  if (mnemonicMatches) {
    return null;
  }

  return (
    <Alert>
      <TriangleAlertIcon className="h-4 w-4" />
      <AlertTitle>
        Your Cashu wallet uses a different recovery phrase
      </AlertTitle>
      <AlertDescription>
        <p>
          Please send your funds to a different wallet, then go to settings{" "}
          {"->"} debug tools {"->"} execute node command {"->"}{" "}
          <span className="font-mono">reset</span>. You will then receive a
          fresh cashu wallet with the correct recovery phrase.
        </p>
      </AlertDescription>
    </Alert>
  );
}
