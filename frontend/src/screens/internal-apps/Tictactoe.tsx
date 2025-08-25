import { AlertTriangleIcon, CopyIcon, ExternalLinkIcon } from "lucide-react";
import React, { useEffect } from "react";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import QRCode from "src/components/QRCode";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { useApps } from "src/hooks/useApps";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";
import { openLink } from "src/utils/openLink";

export function Tictactoe() {
  const appId = "tictactoe";
  const [isLoading, setLoading] = React.useState(false);
  const [appLink, setAppLink] = React.useState("");
  const { data: appsData, mutate: reloadApps } = useApps(undefined, undefined, {
    appStoreAppId: appId,
  });
  const tictactoeApps = appsData?.apps;
  const appStoreApp = appStoreApps.find((app) => app.id === appId)!;

  useEffect(() => {
    if (appLink) {
      window.open(appLink, "_blank");
    }
  }, [appLink]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        const createAppResponse = await createApp({
          name: appStoreApp.title,
          scopes: ["get_info", "lookup_invoice", "make_invoice", "pay_invoice"],
          maxAmount: 30_000,
          budgetRenewal: "monthly",
          metadata: {
            app_store_app_id: appId,
          },
        });

        setAppLink(
          `https://lntictactoe.com/#nwc=${encodeURIComponent(createAppResponse.pairingUri)}`
        );
        toast("Tic Tac Toe connection created");
      } catch (error) {
        handleRequestError("Failed to create connection", error);
      }
      setLoading(false);
      reloadApps();
    })();
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title={
          <div className="flex flex-row items-center">
            <img src={appStoreApp.logo} className="w-14 h-14 rounded-lg mr-4" />
            <div className="flex flex-col">
              <div>{appStoreApp.title}</div>
              <div className="text-sm font-normal text-muted-foreground">
                {appStoreApp.description}
              </div>
            </div>
          </div>
        }
      />
      {appLink ? (
        <div className="max-w-lg flex flex-col gap-5">
          <p>Open the link below to start playing.</p>
          <Alert>
            <AlertTriangleIcon />
            <AlertTitle>
              Save this link and add it to your home screen
            </AlertTitle>
            <AlertDescription>
              This link will only be shown once and can't be retrieved
              afterwards. Please make sure to keep it somewhere safe.
            </AlertDescription>
          </Alert>
          <div
            className="flex flex-col items-center relative cursor-pointer"
            onClick={() => openLink(appLink)}
          >
            <QRCode value={appLink} />
            <img
              src={appStoreApp.logo}
              className="absolute w-12 h-12 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-muted p-1 rounded-xl"
            />
          </div>
          <div className="flex gap-2">
            <Input disabled readOnly type="text" value={appLink} />
            <Button onClick={() => copyToClipboard(appLink)} variant="outline">
              <CopyIcon />
              Copy
            </Button>
            <Button onClick={() => openLink(appLink)} variant="outline">
              <ExternalLinkIcon />
              Open
            </Button>
          </div>
        </div>
      ) : (
        <Card className="max-w-lg">
          <CardHeader>
            <CardTitle className="text-2xl">About the App</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <p className="text-muted-foreground">
              By connecting Tic Tac Toe to your Alby Hub, you can play
              <br />
              tic tac toe with your friends and earn satoshis.
            </p>
            <div className="flex flex-col gap-5">
              <form
                onSubmit={handleSubmit}
                className="flex flex-col items-start gap-5 max-w-lg"
              >
                <LoadingButton loading={isLoading} type="submit">
                  Start playing
                </LoadingButton>
              </form>
            </div>
          </CardContent>
        </Card>
      )}
      {!!tictactoeApps?.length && (
        <>
          <h2 className="font-semibold text-xl">Tic Tac Toe connections</h2>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
            {tictactoeApps.map((app, index) => (
              <AppCard key={index} app={app} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
