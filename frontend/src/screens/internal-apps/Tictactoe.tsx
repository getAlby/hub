import React, { useEffect } from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import { suggestedApps } from "src/components/SuggestedAppData";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function Tictactoe() {
  const appId = "tictactoe";
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const { data: appsData, mutate: reloadApps } = useApps(undefined, undefined, {
    appStoreAppId: appId,
  });
  const tictactoeApps = appsData?.apps;
  const app = suggestedApps.find((app) => app.id === appId)!;

  let appLink: string | undefined;
  if (connectionSecret) {
    appLink = `https://lntictactoe.com/#nwc=${encodeURIComponent(connectionSecret)}`;
  }

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
          name: app.title,
          scopes: ["get_info", "lookup_invoice", "make_invoice", "pay_invoice"],
          maxAmount: 30_000,
          budgetRenewal: "monthly",
          metadata: {
            app_store_app_id: appId,
          },
        });

        setConnectionSecret(createAppResponse.pairingUri);

        toast({ title: "Tic Tac Toe connection created" });
      } catch (error) {
        handleRequestError(toast, "Failed to create connection", error);
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
            <img src={app.logo} className="w-14 h-14 rounded-lg mr-4" />
            <div className="flex flex-col">
              <div>{app.title}</div>
              <div className="text-sm font-normal text-muted-foreground">
                {app.description}
              </div>
            </div>
          </div>
        }
      />
      {appLink ? (
        <div className="max-w-lg flex flex-col gap-5">
          <p>
            Open{" "}
            <a href={appLink} target="_blank" className="underline">
              {app.title}
            </a>{" "}
            to start playing.
          </p>
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
