import { CheckIcon, CopyIcon, ExternalLinkIcon, EyeIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";

import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { IsolatedAppTopupDialog } from "src/components/IsolatedAppTopupDialog";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import {
  SuggestedApp,
  suggestedApps,
} from "src/components/connections/SuggestedAppData";
import { Badge } from "src/components/ui/badge";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useAppByPubkey } from "src/hooks/useApp";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import { App, CreateAppResponse } from "src/types";

export default function AppCreated() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const createAppResponse = state as CreateAppResponse | undefined;
  if (!createAppResponse?.pairingUri) {
    navigate("/");
    return null;
  }

  return <AppCreatedInternal />;
}
function AppCreatedInternal() {
  const { search, state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();

  const queryParams = new URLSearchParams(search);
  const appId = queryParams.get("app") ?? "";
  const appstoreApp = suggestedApps.find((app) => app.id === appId);

  const createAppResponse = state as CreateAppResponse;

  const pairingUri = createAppResponse.pairingUri;
  const { data: app } = useAppByPubkey(
    createAppResponse.pairingPublicKey,
    true
  );

  useEffect(() => {
    if (app?.lastUsedAt) {
      toast({
        title: "Connection established!",
        description: "You can now use the app with your Alby Hub.",
      });
      navigate("/apps");
    }
  }, [app?.lastUsedAt, navigate, toast]);

  useEffect(() => {
    // dispatch a success event which can be listened to by the opener or by the app that embedded the webview
    // this gives those apps the chance to know the user has enabled the connection
    const nwcEvent = new CustomEvent("nwc:success", {
      detail: {
        relayUrl: createAppResponse.relayUrl,
        walletPubkey: createAppResponse.walletPubkey,
        lud16: createAppResponse.lud16,
      },
    });
    window.dispatchEvent(nwcEvent);

    // notify the opener of the successful connection
    if (window.opener) {
      window.opener.postMessage(
        {
          type: "nwc:success",
          relayUrl: createAppResponse.relayUrl,
          walletPubkey: createAppResponse.walletPubkey,
          lud16: createAppResponse.lud16,
        },
        "*"
      );
    }
  }, [
    createAppResponse.relayUrl,
    createAppResponse.walletPubkey,
    createAppResponse.lud16,
  ]);

  if (!createAppResponse) {
    return <Navigate to="/apps/new" />;
  }

  return (
    <>
      <AppHeader
        title={`Connect to ${createAppResponse.name}`}
        description="Configure wallet permissions for the app and follow instructions to finalize the connection"
      />
      <div className="flex flex-col gap-3 sensitive">
        <div>
          <ol className="list-decimal list-inside">
            <li>
              Open{" "}
              {appstoreApp?.webLink ? (
                <ExternalLink
                  className="font-semibold underline"
                  to={appstoreApp.webLink}
                >
                  {appstoreApp.title}
                </ExternalLink>
              ) : (
                "the app you wish to connect"
              )}{" "}
              and look for a way to attach a wallet (most apps provide this
              option in settings)
            </li>
            {app?.isolated && (
              <li>
                Optional: Top up sub-wallet balance (
                {new Intl.NumberFormat().format(Math.floor(app.balance / 1000))}{" "}
                sats){" "}
                <IsolatedAppTopupDialog appPubkey={app.appPubkey}>
                  <Button size="sm" variant="secondary">
                    Top Up
                  </Button>
                </IsolatedAppTopupDialog>
              </li>
            )}
            <li>Scan or paste the connection secret</li>
          </ol>
        </div>
        {app && (
          <ConnectAppCard
            app={app}
            pairingUri={pairingUri}
            appstoreApp={appstoreApp}
          />
        )}
      </div>
    </>
  );
}

export function ConnectAppCard({
  app,
  pairingUri,
  appstoreApp,
}: {
  app: App;
  pairingUri: string;
  appstoreApp?: SuggestedApp;
}) {
  const [timeout, setTimeout] = useState(false);
  const [isQRCodeVisible, setIsQRCodeVisible] = useState(false);
  const { toast } = useToast();
  const copy = () => {
    copyToClipboard(pairingUri, toast);
  };

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      setTimeout(true);
    }, 30000);

    return () => window.clearTimeout(timeoutId);
  }, []);

  return (
    <Card className="max-w-sm">
      <CardHeader>
        <CardTitle className="text-center">Connection Secret</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-5">
        {!app.lastUsedAt ? (
          <>
            <div className="flex flex-row items-center gap-2 text-sm">
              <Loading className="w-4 h-4" />
              <p>Waiting for app to connect</p>
            </div>
            {timeout && (
              <div className="text-sm flex flex-col gap-2 items-center text-center">
                Connecting is taking longer than usual.
                <Link to={`/apps/${app?.appPubkey}`}>
                  <Button variant="secondary">Continue anyway</Button>
                </Link>
              </div>
            )}
          </>
        ) : (
          <Badge variant="positive">
            <CheckIcon className="w-4 h-4 mr-2" />
            <p>App connected</p>
          </Badge>
        )}
        <div className="relative">
          <div
            className={cn(!isQRCodeVisible && "blur-md cursor-pointer")}
            onClick={() => setIsQRCodeVisible(true)}
          >
            <QRCode className={"w-full"} value={pairingUri} />
            {appstoreApp && (
              <img
                src={appstoreApp.logo}
                className="absolute w-12 h-12 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-muted p-1 rounded-xl"
              />
            )}
          </div>
          {!isQRCodeVisible && (
            <Button
              onClick={() => {
                setIsQRCodeVisible(true);
              }}
              className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
            >
              <EyeIcon className="h-4 w-4 mr-2" />
              Reveal QR
            </Button>
          )}
        </div>
        <div className="flex gap-2">
          <Button onClick={copy} variant="outline">
            <CopyIcon className="w-4 h-4 mr-2" />
            Copy
          </Button>
          <ExternalLinkButton to={pairingUri} variant="outline">
            <ExternalLinkIcon className="w-4 h-4 mr-2" />
            Open
          </ExternalLinkButton>
        </div>
      </CardContent>
    </Card>
  );
}
