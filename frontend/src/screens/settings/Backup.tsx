import {
  CopyIcon,
  ExternalLinkIcon,
  Link2Icon,
  TriangleAlertIcon,
} from "lucide-react";
import React, { useState } from "react";

import { useNavigate } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import MnemonicInputs from "src/components/MnemonicInputs";
import PasswordInput from "src/components/password/PasswordInput";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
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
import { Checkbox } from "src/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Separator } from "src/components/ui/separator";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { useMigrateLDKStorage } from "src/hooks/useMigrateLDKStorage";
import { copyToClipboard } from "src/lib/clipboard";
import { MnemonicResponse } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request";

export default function Backup() {
  const { toast } = useToast();
  const { data: info, hasMnemonic } = useInfo();
  const { data: me } = useAlbyMe();
  const { isMigratingStorage, migrateLDKStorage } = useMigrateLDKStorage();
  const [unlockPassword, setUnlockPassword] = useState("");
  const [decryptedMnemonic, setDecryptedMnemonic] = useState("");
  const [loading, setLoading] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const navigate = useNavigate();
  const { mutate: refetchInfo } = useInfo();
  const [backedUp, setIsBackedUp] = useState<boolean>(false);
  const [backedUp2, setIsBackedUp2] = useState<boolean>(false);

  const onSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      const result = await request<MnemonicResponse>("/api/mnemonic", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ unlockPassword }),
      });

      setDecryptedMnemonic(result?.mnemonic ?? "");
      setIsDialogOpen(true);
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

    const sixMonthsLater = new Date();
    sixMonthsLater.setMonth(sixMonthsLater.getMonth() + 6);

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
        title="Backup"
        description={
          <>
            <span className="text-muted-foreground">
              Backup your wallet recovery phrase and channel states. These
              backups are for disaster recovery only. To migrate your node,
              please use the migration tool.
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
              Your recovery phrase is a group of 12 random words that back up
              your wallet on-chain balance. Using them is the only way to
              recover access to your wallet on another machine or when you loose
              your unlock password.
            </p>
          </div>
          <p className="text-destructive">
            If you loose access to your Hub and do not have your recovery
            phrase, you will loose access to your funds.
          </p>
          {info?.backendType === "CASHU" && <CashuMnemonicWarning />}

          <div>
            <form
              onSubmit={onSubmitPassword}
              className="max-w-md flex flex-col gap-3"
            >
              <div className="grid gap-2 mb-6">
                <Label htmlFor="password">Password</Label>
                <PasswordInput id="password" onChange={setUnlockPassword} />
                <p className="text-sm text-muted-foreground">
                  Enter your unlock password to view your recovery phrase.
                </p>
              </div>
              <div className="flex justify-start">
                <LoadingButton loading={loading} variant="secondary" size="lg">
                  View Recovery Phrase
                </LoadingButton>
              </div>
            </form>
          </div>

          <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Wallet Recovery Phrase</DialogTitle>
                <DialogDescription>
                  Write these words down, store them somewhere safe, and keep
                  them secret.
                </DialogDescription>
              </DialogHeader>
              <form
                onSubmit={onSubmit}
                className="flex flex-col gap-2 max-w-md text-sm"
              >
                <MnemonicInputs mnemonic={decryptedMnemonic} readOnly={true} />
                <div className="flex justify-center mt-4">
                  <Button
                    type="button"
                    variant={"destructive"}
                    className="flex gap-2 justify-center"
                    onClick={() => copyToClipboard(decryptedMnemonic, toast)}
                  >
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Dangerously Copy
                  </Button>
                </div>
                <div className="flex items-center mt-6 text-sm">
                  <Checkbox
                    id="backup"
                    required
                    onCheckedChange={() => setIsBackedUp(!backedUp)}
                  />
                  <Label htmlFor="backup" className="ml-2">
                    I've backed up my recovery phrase to my wallet in a private
                    and secure place
                  </Label>
                </div>
                {backedUp && !info?.albyAccountConnected && (
                  <div className="flex text-sm">
                    <Checkbox
                      id="backup2"
                      required
                      onCheckedChange={() => setIsBackedUp2(!backedUp2)}
                    />
                    <Label
                      htmlFor="backup2"
                      className="ml-2 text-sm text-foreground"
                    >
                      I understand the recovery phrase AND a backup of my hub
                      data directory is required to recover funds from my
                      lightning channels.
                    </Label>
                  </div>
                )}
                <div className="flex justify-end gap-2 items-center">
                  <div className="flex justify-center">
                    <Button type="submit" size="lg">
                      Finish Backup
                    </Button>
                  </div>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      )}
      <>
        <Separator className="my-6" />
        <div className="flex flex-col gap-8">
          <div>
            <h3 className="text-lg font-medium">Channels Backup</h3>
            <p className="text-sm text-muted-foreground">
              Your spending balance is stored in your lightning channels. In
              case of recovery of your Alby Hub, they need to be backed up every
              time you open a new channel.
            </p>
          </div>

          <div>
            <div className="flex gap-2 mb-1 items-center">
              <h3 className="text-sm font-medium">Automated Channels Backup</h3>
              {info?.albyAccountConnected ? (
                <Badge variant={"positive"}>Active</Badge>
              ) : (
                <Badge>Recommended</Badge>
              )}
            </div>
            {info?.albyAccountConnected ? (
              <>
                <p className="text-muted-foreground text-sm mb-8">
                  Your channel state is backed up automatically after each
                  channel creation. Potential recovery will trigger channel
                  closures, and your funds will arrive in your on-chain balance.
                </p>
                {info?.vssSupported && (
                  <>
                    <div>
                      <div className="flex gap-2 mb-1 items-center">
                        <h3 className="text-sm font-medium">
                          Dynamic Channels Backup With Instant Recovery
                        </h3>
                        {me?.subscription.buzz && info.ldkVssEnabled ? (
                          <Badge variant={"positive"}>Active</Badge>
                        ) : (
                          <Badge>Alby Cloud</Badge>
                        )}
                      </div>
                      <p className="text-sm text-muted-foreground mb-4">
                        When enabled, your channels state is dynamically updated
                        and stored end-to-end encrypted by Alby's Versioned
                        Storage Service. This allows you to recover your
                        spending balance with your recovery phrase alone,
                        without having to close your channels.
                      </p>

                      {!info.ldkVssEnabled &&
                        (!me?.subscription.buzz ? (
                          <Button
                            variant="secondary"
                            disabled={info.ldkVssEnabled}
                            size={"lg"}
                            onClick={() => {
                              if (!me?.subscription?.buzz) {
                                openLink(
                                  "https://getalby.com/subscription/new"
                                );
                              }
                            }}
                          >
                            Enable Dynamic Channels Backup
                          </Button>
                        ) : (
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <LoadingButton
                                variant="secondary"
                                loading={isMigratingStorage}
                                disabled={info.ldkVssEnabled}
                                size={"lg"}
                              >
                                Enable Dynamic Channels Backup
                              </LoadingButton>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogHeader>
                                <AlertDialogTitle>
                                  Alby Hub Restart Required
                                </AlertDialogTitle>
                                <AlertDialogDescription>
                                  <div>
                                    <p>
                                      As part of enabling VSS your hub will be
                                      shut down, and you will need to enter your
                                      unlock password to start it again.
                                    </p>
                                    <p className="mt-2">
                                      Please ensure you have no pending payments
                                      or channel closures before continuing.
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
                        ))}
                    </div>
                  </>
                )}
              </>
            ) : (
              <div className="flex flex-col gap-8">
                <div>
                  <p className="text-muted-foreground text-sm mb-4">
                    Link your Alby Account to enable automated channel backups
                    after each channel creation.
                  </p>
                  <Button
                    type="button"
                    variant={"secondary"}
                    className="flex gap-2 justify-center"
                    onClick={() => navigate("/alby/account")}
                  >
                    <Link2Icon className="w-4 h-4 mr-2" />
                    Link Alby Account to Enable
                  </Button>
                </div>
                <div className="flex flex-col gap-4">
                  <div className="flex gap-2 mb-1 items-center">
                    <h3 className="text-sm font-medium">
                      Manual Channels Backup
                    </h3>
                    <Badge variant={"positive"}>Active</Badge>
                  </div>
                  <p className="text-muted-foreground text-sm ">
                    To backup your channels state manually, without Alby Account
                    linked, follow these steps:
                  </p>
                  <ol className="text-sm list-decimal list-inside">
                    <li className="mb-1">
                      Go to the working directory of your Alby Hub
                    </li>
                    <li>
                      Back up the newest file in: Set your Unlock Passcode{" "}
                      <code className="bg-muted rounded-lg py-1 px-1">
                        .data/ldk/static_channel_backups/
                      </code>
                    </li>
                  </ol>
                  <p className="text-destructive text-sm">
                    If new channels are created after the backup, you could risk
                    losing funds.
                  </p>
                  <ExternalLink
                    to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/backups-and-recover#alby-hub-self-hosted-without-an-alby-account"
                    className="underline flex items-center mt-4"
                  >
                    View manual backups guide
                    <ExternalLinkIcon className="w-4 h-4 ml-2" />
                  </ExternalLink>
                </div>
              </div>
            )}
          </div>
        </div>
      </>

      {!hasMnemonic && !info?.vssSupported && (
        <p className="text-sm text-muted-foreground">
          No wallet recovery phrase or channel state backup present.
        </p>
      )}
    </>
  );
}

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
