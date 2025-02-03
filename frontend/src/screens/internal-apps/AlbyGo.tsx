import { Globe, InfoIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { ChromeIcon } from "src/components/icons/Chrome";
import { FirefoxIcon } from "src/components/icons/Firefox";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { ZapStoreIcon } from "src/components/icons/ZapStore";
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
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { useApp } from "src/hooks/useApp";
import { createApp } from "src/requests/createApp";
import { ConnectAppCard } from "src/screens/apps/AppCreated";

export function AlbyGo() {
  const [isSuperuser, setSuperuser] = React.useState(true);
  const [loading, setLoading] = React.useState(false);
  const [appPubkey, setAppPubkey] = React.useState<string>();
  const [connectionSecret, setConnectionSecret] = React.useState<string>("");
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [showCreateConnectionDialog, setShowCreateConnectionDialog] =
    React.useState(false);
  const app = suggestedApps.find((app) => app.id === "alby-go");
  const { data: createdApp } = useApp(appPubkey, true);
  const navigate = useNavigate();
  const { toast } = useToast();
  if (!app) {
    return null;
  }

  function onClickCreateConnection() {
    if (!isSuperuser) {
      navigate("/apps/new?app=alby-go");
      return;
    }
    setShowCreateConnectionDialog(true);
  }
  async function onSubmitCreateConnection(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      const createAppResponse = await createApp({
        name: "Alby Go",
        scopes: [
          "pay_invoice",
          "get_balance",
          "get_info",
          "make_invoice",
          "lookup_invoice",
          "list_transactions",
          "sign_message",
          "notifications",
          "superuser",
        ],
        isolated: false,
        metadata: {
          app_store_app_id: "alby-go",
        },
        unlockPassword,
      });
      setConnectionSecret(createAppResponse.pairingUri);
      setAppPubkey(createAppResponse.pairingPublicKey);
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
  }

  return (
    <div className="grid gap-5">
      {showCreateConnectionDialog && (
        <AlertDialog open={showCreateConnectionDialog}>
          <AlertDialogContent>
            <form onSubmit={onSubmitCreateConnection}>
              <AlertDialogHeader>
                <AlertDialogTitle>Confirm New Connection</AlertDialogTitle>
                <AlertDialogDescription>
                  <div>
                    <p>
                      Alby Go will be given permission to create other app
                      connections which can spend your balance. Please enter
                      your unlock password to continue.
                    </p>
                    <div className="grid gap-1.5 mt-2">
                      <Label htmlFor="password">Unlock Password</Label>
                      <Input
                        autoFocus
                        type="password"
                        name="password"
                        onChange={(e) => setUnlockPassword(e.target.value)}
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
      )}
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
          {createdApp && connectionSecret && (
            <ConnectAppCard app={createdApp} pairingUri={connectionSecret} />
          )}
          {!createdApp && (
            <Card>
              <CardHeader>
                <CardTitle className="text-2xl">Configure Alby Go</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center mt-2">
                  <Checkbox
                    id="superuser"
                    className="mr-2"
                    onCheckedChange={(e) =>
                      setSuperuser(e.valueOf() as boolean)
                    }
                    checked={isSuperuser}
                  />
                  <Label htmlFor="superuser" className="cursor-pointer">
                    Allow Alby Go to create other app connections{" "}
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger>
                          <div className="flex flex-row gap-1 items-center text-muted-foreground">
                            <InfoIcon className="h-3 w-3 shrink-0" />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent className="w-[400px]">
                          Enable other mobile apps to quickly connect to your
                          hub by confirming within Alby Go. Please be aware that
                          any budget set on Alby Go will not apply to any newly
                          created apps.
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </Label>
                </div>
                {
                  <Button className="mt-8" onClick={onClickCreateConnection}>
                    <NostrWalletConnectIcon className="w-4 h-4 mr-2" />
                    Create App Connection
                  </Button>
                }
              </CardContent>
            </Card>
          )}
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
                      <PlayStoreIcon className="w-4 h-4 mr-2" />
                      Play Store
                    </Button>
                  </ExternalLink>
                )}
                {app.appleLink && (
                  <ExternalLink to={app.appleLink}>
                    <Button variant="outline">
                      <AppleIcon className="w-4 h-4 mr-2" />
                      App Store
                    </Button>
                  </ExternalLink>
                )}
                {app.zapStoreLink && (
                  <ExternalLink to={app.zapStoreLink}>
                    <Button variant="outline">
                      <ZapStoreIcon className="w-4 h-4 mr-2" />
                      Zapstore
                    </Button>
                  </ExternalLink>
                )}
                {app.chromeLink && (
                  <ExternalLink to={app.chromeLink}>
                    <Button variant="outline">
                      <ChromeIcon className="w-4 h-4 mr-2" />
                      Chrome Web Store
                    </Button>
                  </ExternalLink>
                )}
                {app.firefoxLink && (
                  <ExternalLink to={app.firefoxLink}>
                    <Button variant="outline">
                      <FirefoxIcon className="w-4 h-4 mr-2" />
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
                      <Globe className="w-4 h-4 mr-2" />
                      Website
                    </Button>
                  </ExternalLink>
                )}
              </CardFooter>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
