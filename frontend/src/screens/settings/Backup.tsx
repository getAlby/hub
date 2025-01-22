import React from "react";
import { Link, useNavigate } from "react-router-dom";
import PasswordViewAdornment from "src/components/PasswordAdornment";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert } from "src/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Separator } from "src/components/ui/separator";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { useMigrateLDKStorage } from "src/hooks/useMigrateLDKStorage";
import { MnemonicResponse } from "src/types";
import { request } from "src/utils/request";

export default function Backup() {
  const navigate = useNavigate();

  const { toast } = useToast();
  const { data: info, hasNodeBackup, hasMnemonic } = useInfo();
  const { data: me } = useAlbyMe();
  const { isMigratingStorage, migrateLDKStorage } = useMigrateLDKStorage();
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [unlockPasswordVisible, setUnlockPasswordVisible] =
    React.useState(false);

  const [loading, setLoading] = React.useState(false);

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

      if (result?.mnemonic) {
        navigate("/settings/mnemonic-backup");
      }
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

  return (
    <>
      <SettingsHeader
        title="Backup"
        description={
          <>
            <span className="text-muted-foreground">
              Backup your wallet recovery phrase and or your channel states in
              order to migrate your node.{" "}
            </span>
            <a
              href="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/backups-and-recover"
              target="_blank"
              rel="noreferrer noopener"
              className="text-foreground underline"
            >
              {" "}
              Learn more about backups
            </a>
          </>
        }
      />

      {hasMnemonic && (
        <div className="flex flex-col gap-6">
          <div>
            <h3 className="text-lg font-medium">Wallet Keys Backup</h3>
            <p className="text-sm text-muted-foreground">
              Key recovery phrase is a group of 12 random words that back up
              your wallet on-chain balance. Using them is the only way to
              recover access to your wallet on another machine or when you loose
              your unlock password. .&nbsp;
              {info?.albyAccountConnected && (
                <>
                  Channel backups are saved automatically to your Alby Account,
                  encrypted with your recovery phrase.
                </>
              )}
              {!info?.albyAccountConnected && (
                <>
                  Make sure to also backup your <b>data directory</b> as this is
                  required to recover funds on your channels. You can also
                  connect your Alby Account for automatic encrypted backups.
                </>
              )}
            </p>
          </div>
          <p className="text-destructive">
            If you loose access to your Hub and do not have your recovery
            phrase, you will loose access to your funds.
          </p>
          <form
            onSubmit={onSubmitPassword}
            className="max-w-md flex flex-col gap-3"
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
              <p className="text-sm text-muted-foreground">
                Enter your unlock password to view your recovery phrase.
              </p>
            </div>
            <div className="flex justify-start">
              <LoadingButton
                loading={loading}
                variant="secondary"
                size={"lg"}
                disabled={!unlockPassword}
              >
                View Recovery Phase
              </LoadingButton>
            </div>
          </form>
        </div>
      )}

      {(info?.vssSupported || hasNodeBackup) && (
        <>
          <Separator className="my-6" />
          <div className="flex flex-col gap-4">
            <h3 className="text-lg font-medium">Channels Backup</h3>

            {info?.vssSupported && (
              <>
                <div>
                  <div className="flex gap-2 mb-1 items-center">
                    <h3 className="text-sm font-medium">
                      Automated Channels Backups
                    </h3>
                    <Badge className="ml-2">
                      {me?.subscription.buzz && info.ldkVssEnabled
                        ? "Premium"
                        : "Alby Cloud"}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mb-4">
                    When enabled, channels backups are saved automatically a
                    virtual disk encrypted with your wallet recovery phrase
                    thanks to Versioned Storage Service (VSS). This allows you
                    to recover or migrate your Hub without having to close your
                    channels.
                  </p>

                  {me?.subscription.buzz && (
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        {
                          <LoadingButton
                            variant="secondary"
                            loading={isMigratingStorage}
                            disabled={info.ldkVssEnabled}
                            size={"lg"}
                          >
                            {info.ldkVssEnabled
                              ? "Automated Backups Enabled"
                              : "Enable Automated Backups"}
                          </LoadingButton>
                        }
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>
                            Alby Hub Restart Required
                          </AlertDialogTitle>
                          <AlertDialogDescription>
                            <div>
                              <p>
                                As part of enabling VSS your hub will be shut
                                down, and you will need to enter your unlock
                                password to start it again.
                              </p>
                              <p className="mt-2">
                                Please ensure you have no pending payments or
                                channel closures before continuing.
                              </p>
                            </div>
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            onClick={() => migrateLDKStorage("VSS")}
                          >
                            Confirm
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  )}

                  {!me?.subscription.buzz && (
                    <Alert variant={"default"}>
                      VSS is only available to Alby users with a paid
                      subscription.
                    </Alert>
                  )}
                </div>
              </>
            )}

            {hasNodeBackup && (
              <div>
                <h3 className="text-sm font-medium mb-1">Migrate Alby Hub</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  If you’d like to import or migrate your Hub onto another
                  device or server, you’ll need your channels’ backup file to
                  import your channels state. This instance of Hub will be
                  stopped.
                </p>
                <Link to="/settings/node-backup">
                  <Button variant="secondary" size={"lg"}>
                    Migrate Your Alby Hub
                  </Button>
                </Link>
              </div>
            )}
          </div>
        </>
      )}

      {!hasMnemonic && !hasNodeBackup && !info?.vssSupported && (
        <p className="text-sm text-muted-foreground">
          No wallet recovery phrase or channel state backup present.
        </p>
      )}
    </>
  );
}
