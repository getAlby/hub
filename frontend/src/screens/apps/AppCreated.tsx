import { CopyIcon, EyeIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";

import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { suggestedApps } from "src/components/SuggestedAppData";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useApp } from "src/hooks/useApp";
import { copyToClipboard } from "src/lib/clipboard";
import { CreateAppResponse } from "src/types";

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

  const [timeout, setTimeout] = useState(false);
  const [isQRCodeVisible, setIsQRCodeVisible] = useState(false);

  const createAppResponse = state as CreateAppResponse;
  const pairingUri = createAppResponse.pairingUri;
  const { data: app } = useApp(createAppResponse.pairingPublicKey, true);

  const copy = () => {
    copyToClipboard(pairingUri);
    toast({ title: "Copied to clipboard." });
  };

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      setTimeout(true);
    }, 30000);

    return () => window.clearTimeout(timeoutId);
  }, []);

  useEffect(() => {
    if (app?.lastEventAt) {
      toast({
        title: "Connection established!",
        description: "You can now use the app with your Alby Hub.",
      });
      navigate("/apps");
    }
  }, [app?.lastEventAt, navigate, toast]);

  useEffect(() => {
    if (appstoreApp) {
      return;
    }
    // dispatch a success event which can be listened to by the opener or by the app that embedded the webview
    // this gives those apps the chance to know the user has enabled the connection
    const nwcEvent = new CustomEvent("nwc:success", { detail: {} });
    window.dispatchEvent(nwcEvent);

    // notify the opener of the successful connection
    if (window.opener) {
      window.opener.postMessage(
        {
          type: "nwc:success",
          payload: { success: true },
        },
        "*"
      );
    }
  }, [appstoreApp]);

  if (!createAppResponse) {
    return <Navigate to="/apps/new" />;
  }

  return (
    <>
      <AppHeader
        title={`Connect to ${createAppResponse.name}`}
        description="Configure wallet permissions for the app and follow instructions to finalise the connection"
      />
      <div className="flex flex-col gap-3 sensitive">
        <div>
          <p>
            1. Open{" "}
            {appstoreApp ? (
              <ExternalLink
                className="font-semibold underline"
                to={appstoreApp.webLink}
              >
                {appstoreApp.title}
              </ExternalLink>
            ) : (
              "the app you wish to connect"
            )}{" "}
            and look for a way to attach a wallet (most apps provide this option
            in settings)
          </p>
          <p>2. Scan or paste the connection secret</p>
        </div>
        <Card className="max-w-sm">
          <CardHeader>
            <CardTitle className="text-center">Connection Secret</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-5">
            <div className="flex flex-row items-center gap-2 text-sm">
              <Loading className="w-4 h-4" />
              <p>Waiting for app to connect</p>
            </div>
            {timeout && (
              <div className="text-sm flex flex-col gap-2 items-center text-center">
                Connecting is taking longer than usual.
                <Link to={`/apps/${app?.nostrPubkey}`}>
                  <Button variant="secondary">Continue anyway</Button>
                </Link>
              </div>
            )}
            <a href={pairingUri} target="_blank" className="relative">
              <div className={!isQRCodeVisible ? "blur-md" : ""}>
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
                  onClick={(e) => {
                    e.preventDefault();
                    setIsQRCodeVisible(true);
                  }}
                  className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
                >
                  <EyeIcon className="h-4 w-4 mr-2" />
                  Reveal QR
                </Button>
              )}
            </a>
            <div>
              <Button onClick={copy} variant="outline">
                <CopyIcon className="w-4 h-4 mr-2" />
                Copy pairing secret
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
