import { GlobeIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import { AppDetailConnectedApps } from "src/components/connections/AppDetailConnectedApps";
import { AppDetailHeader } from "src/components/connections/AppDetailHeader";
import { suggestedApps } from "src/components/connections/SuggestedAppData";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { ChromeIcon } from "src/components/icons/Chrome";
import { FirefoxIcon } from "src/components/icons/Firefox";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { ZapStoreIcon } from "src/components/icons/ZapStore";
import Loading from "src/components/Loading";
import PasswordInput from "src/components/password/PasswordInput";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useCapabilities } from "src/hooks/useCapabilities";
import { createApp } from "src/requests/createApp";

export function AlbyGo() {
  const [loading, setLoading] = React.useState(false);
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [showCreateConnectionDialog, setShowCreateConnectionDialog] =
    React.useState(false);
  const { toast } = useToast();
  const { data: capabilities } = useCapabilities();
  const navigate = useNavigate();

  const appStoreApp = suggestedApps.find((app) => app.id === "alby-go");
  if (!appStoreApp) {
    return null;
  }

  function onClickCreateConnection() {
    setShowCreateConnectionDialog(true);
  }

  async function onSubmitCreateConnection(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      if (!appStoreApp) {
        throw new Error("Alby go app not found");
      }
      if (!capabilities) {
        throw new Error("capabilities not loaded");
      }

      // TODO: fetch scopes from useCapabilities
      const createAppResponse = await createApp({
        name: "Alby Go",
        scopes: [...capabilities.scopes, "superuser"],
        isolated: false,
        metadata: {
          app_store_app_id: appStoreApp.id,
        },
        unlockPassword,
        maxAmount: 100_000,
        budgetRenewal: "monthly",
      });
      navigate(`/apps/created?app=${appStoreApp.id}`, {
        state: createAppResponse,
      });
      toast({ title: "Alby Go connection created" });
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Something went wrong: " + error,
      });
    }
    setLoading(false);
    setShowCreateConnectionDialog(false);
    setUnlockPassword("");
  }

  if (!capabilities) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AlertDialog open={showCreateConnectionDialog}>
        <AlertDialogContent>
          <form onSubmit={onSubmitCreateConnection}>
            <AlertDialogHeader>
              <AlertDialogTitle>Confirm New Connection</AlertDialogTitle>
              <AlertDialogDescription>
                <div className="flex flex-col">
                  <p>
                    Alby Go will be given permission to create other app
                    connections which can spend your balance.
                  </p>

                  <p className="mt-4">
                    Alby Go will be given a 100k sat / month budget by default
                    which you can edit after creating the connection.
                  </p>

                  <p className="mt-4">
                    Warning: Alby Go can create connections with a larger budget
                    than the one set for Alby Go. Make sure to always set a
                    budget.
                  </p>

                  <p className="mt-4">
                    Please enter your unlock password to continue.
                  </p>
                  <div className="grid gap-1.5 mt-2">
                    <Label htmlFor="password">Unlock Password</Label>
                    <PasswordInput
                      id="password"
                      onChange={setUnlockPassword}
                      autoFocus
                      value={unlockPassword}
                    />
                  </div>
                </div>
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter className="mt-3">
              <AlertDialogCancel
                onClick={() => setShowCreateConnectionDialog(false)}
              >
                Cancel
              </AlertDialogCancel>
              <LoadingButton loading={loading} type="submit">
                Confirm
              </LoadingButton>
            </AlertDialogFooter>
          </form>
        </AlertDialogContent>
      </AlertDialog>
      <AppDetailHeader appStoreApp={appStoreApp} />

      <AppDetailConnectedApps appStoreApp={appStoreApp} />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="flex flex-col w-full gap-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">About the App</CardTitle>
            </CardHeader>
            {appStoreApp.extendedDescription && (
              <CardContent className="flex flex-col gap-3">
                <p className="text-muted-foreground">
                  {appStoreApp.extendedDescription}
                </p>
              </CardContent>
            )}
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">How to Connect</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              <>
                <div>
                  <h3 className="font-medium">In Alby Go</h3>
                  <ul className="list-inside text-muted-foreground">
                    <li>
                      1. Download and open{" "}
                      <span className="font-medium text-foreground">
                        Alby Go
                      </span>{" "}
                      on your Android or iOS device
                    </li>
                    <li>
                      2. Click on{" "}
                      <span className="font-medium text-foreground">
                        Connect Wallet
                      </span>
                    </li>
                    <li>
                      3.{" "}
                      <span className="font-medium text-foreground">
                        Scan or paste
                      </span>{" "}
                      the connection secret from Alby Hub that will be revealed
                      once you create the connection below.
                    </li>
                  </ul>
                </div>
              </>
            </CardContent>
          </Card>
        </div>
        <div className="flex flex-col w-full gap-6">
          {(appStoreApp.appleLink ||
            appStoreApp.playLink ||
            appStoreApp.zapStoreLink ||
            appStoreApp.chromeLink ||
            appStoreApp.firefoxLink) && (
            <Card>
              <CardHeader>
                <CardTitle className="text-2xl">Get This App</CardTitle>
              </CardHeader>
              <CardFooter className="flex flex-row gap-2">
                {appStoreApp.playLink && (
                  <ExternalLink to={appStoreApp.playLink}>
                    <Button variant="outline">
                      <PlayStoreIcon className="w-4 h-4 mr-2" />
                      Play Store
                    </Button>
                  </ExternalLink>
                )}
                {appStoreApp.appleLink && (
                  <ExternalLink to={appStoreApp.appleLink}>
                    <Button variant="outline">
                      <AppleIcon className="w-4 h-4 mr-2" />
                      App Store
                    </Button>
                  </ExternalLink>
                )}
                {appStoreApp.zapStoreLink && (
                  <ExternalLink to={appStoreApp.zapStoreLink}>
                    <Button variant="outline">
                      <ZapStoreIcon className="w-4 h-4 mr-2" />
                      Zapstore
                    </Button>
                  </ExternalLink>
                )}
                {appStoreApp.chromeLink && (
                  <ExternalLink to={appStoreApp.chromeLink}>
                    <Button variant="outline">
                      <ChromeIcon className="w-4 h-4 mr-2" />
                      Chrome Web Store
                    </Button>
                  </ExternalLink>
                )}
                {appStoreApp.firefoxLink && (
                  <ExternalLink to={appStoreApp.firefoxLink}>
                    <Button variant="outline">
                      <FirefoxIcon className="w-4 h-4 mr-2" />
                      Firefox Add-Ons
                    </Button>
                  </ExternalLink>
                )}
              </CardFooter>
            </Card>
          )}
          {appStoreApp.webLink && (
            <Card>
              <CardHeader>
                <CardTitle className="text-2xl">Links</CardTitle>
              </CardHeader>
              <CardFooter className="flex flex-row gap-2">
                {appStoreApp.webLink && (
                  <ExternalLink to={appStoreApp.webLink}>
                    <Button variant="outline">
                      <GlobeIcon className="w-4 h-4 mr-2" />
                      Website
                    </Button>
                  </ExternalLink>
                )}
              </CardFooter>
            </Card>
          )}

          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">One Tap Connections</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                Use Alby Go to quickly connect other apps to your hub with one
                tap on mobile.
              </p>
              {
                <Button className="mt-8" onClick={onClickCreateConnection}>
                  <NostrWalletConnectIcon className="w-4 h-4 mr-2" />
                  Connect with One Tap Connections
                </Button>
              }
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
