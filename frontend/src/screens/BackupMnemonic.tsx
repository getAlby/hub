import {
  ExternalLinkIcon,
  LifeBuoy,
  ShieldAlert,
  ShieldCheck,
  TriangleAlertIcon,
} from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

import Container from "src/components/Container";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import MnemonicInputs from "src/components/MnemonicInputs";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
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
        title="Backup Your Keys"
        description="Make sure to keep your backup somewhere safe"
      />
      {info?.backendType === "CASHU" && <CashuMnemonicWarning />}
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
            <div className="flex gap-2 items-center">
              <div className="shrink-0">
                <ShieldCheck className="w-6 h-6" />
              </div>
              <span>
                Make sure to write them down somewhere safe and private.
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

          <MnemonicInputs mnemonic={decryptedMnemonic} readOnly={true}>
            <div className="flex items-center mt-5">
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
              <div className="flex mt-5">
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
