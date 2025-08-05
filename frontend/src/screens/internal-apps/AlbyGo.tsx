import { GlobeIcon } from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { ChromeIcon } from "src/components/icons/Chrome";
import { FirefoxIcon } from "src/components/icons/Firefox";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { ZapStoreIcon } from "src/components/icons/ZapStore";
import Loading from "src/components/Loading";
import PasswordInput from "src/components/password/PasswordInput";
import { suggestedApps } from "src/components/SuggestedAppData";
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

  const app = suggestedApps.find((app) => app.id === "alby-go");
  if (!app) {
    return null;
  }

  function onClickCreateConnection() {
    setShowCreateConnectionDialog(true);
  }

  async function onSubmitCreateConnection(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      if (!app) {
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
          app_store_app_id: app.id,
        },
        unlockPassword,
        maxAmount: 100_000,
        budgetRenewal: "monthly",
      });
      navigate(`/apps/created?app=${app.id}`, {
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
      <AppHeader
        title={
          <>
            <div className="flex flex-row items-center">
              <img src={app.logo} className="w-14 h-14 rounded-lg mr-4" />
              <div className="flex flex-col">
                <div>{app.title}</div>
                <div className="text-sm font-normal text-muted-foreground">
                  {app.description}
                </div>
              </div>
            </div>
          </>
        }
        description=""
        contentRight={
          <Link to={`/apps/new?app=${app.id}`}>
            <Button>
              <NostrWalletConnectIcon className="size-4 mr-2" />
              Connect to {app.title}
            </Button>
          </Link>
        }
      />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="flex flex-col w-full gap-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">About the App</CardTitle>
            </CardHeader>
            {app.extendedDescription && (
              <CardContent className="flex flex-col gap-3">
                <p className="text-muted-foreground">
                  {app.extendedDescription}
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
          {(app.appleLink ||
            app.playLink ||
            app.zapStoreLink ||
            app.chromeLink ||
            app.firefoxLink) && (
            <Card>
              <CardHeader>
                <CardTitle className="text-2xl">Get This App</CardTitle>
              </CardHeader>
              <CardFooter className="flex flex-row gap-2">
                {app.playLink && (
                  <ExternalLink to={app.playLink}>
                    <Button variant="outline">
                      <PlayStoreIcon />
                      Play Store
                    </Button>
                  </ExternalLink>
                )}
                {app.appleLink && (
                  <ExternalLink to={app.appleLink}>
                    <Button variant="outline">
                      <AppleIcon />
                      App Store
                    </Button>
                  </ExternalLink>
                )}
                {app.zapStoreLink && (
                  <ExternalLink to={app.zapStoreLink}>
                    <Button variant="outline">
                      <ZapStoreIcon />
                      Zapstore
                    </Button>
                  </ExternalLink>
                )}
                {app.chromeLink && (
                  <ExternalLink to={app.chromeLink}>
                    <Button variant="outline">
                      <ChromeIcon />
                      Chrome Web Store
                    </Button>
                  </ExternalLink>
                )}
                {app.firefoxLink && (
                  <ExternalLink to={app.firefoxLink}>
                    <Button variant="outline">
                      <FirefoxIcon />
                      Firefox Add-Ons
                    </Button>
                  </ExternalLink>
                )}
              </CardFooter>
            </Card>
          )}
          {app.webLink && (
            <Card>
              <CardHeader>
                <CardTitle className="text-2xl">Links</CardTitle>
              </CardHeader>
              <CardFooter className="flex flex-row gap-2">
                {app.webLink && (
                  <ExternalLink to={app.webLink}>
                    <Button variant="outline">
                      <GlobeIcon />
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
                  <NostrWalletConnectIcon />
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
