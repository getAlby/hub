import { CheckIcon, CopyIcon, EyeIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import { App } from "src/types";

export function ConnectAppCard({
  app,
  pairingUri,
  appStoreApp,
}: {
  app: App;
  pairingUri: string;
  appStoreApp?: AppStoreApp;
}) {
  const [timeout, setTimeout] = useState(false);
  const [isQRCodeVisible, setIsQRCodeVisible] = useState(false);
  const copy = () => {
    copyToClipboard(pairingUri);
  };

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      setTimeout(true);
    }, 30000);

    return () => window.clearTimeout(timeoutId);
  }, []);

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-center">Connection Secret</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-5">
        {!app.lastUsedAt ? (
          <>
            <div className="flex flex-row items-center gap-2 text-sm">
              <Loading className="size-4" />
              <p>Waiting for app to connect</p>
            </div>
            {timeout && (
              <div className="text-sm flex flex-col gap-2 items-center text-center">
                Connecting is taking longer than usual.
                <Link to={`/apps/${app?.id}`}>
                  <Button variant="secondary">Continue anyway</Button>
                </Link>
              </div>
            )}
          </>
        ) : (
          <Badge variant="positive">
            <CheckIcon />
            App connected
          </Badge>
        )}
        {!appStoreApp?.hideConnectionQr && (
          <div className="relative">
            <div
              className={cn(!isQRCodeVisible && "blur-md cursor-pointer")}
              onClick={() => setIsQRCodeVisible(true)}
            >
              <QRCode value={pairingUri} />
              {appStoreApp && (
                <img
                  src={appStoreApp.logo}
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
                <EyeIcon />
                Reveal QR
              </Button>
            )}
          </div>
        )}
        <div className="flex gap-2">
          <Button onClick={copy} variant="outline">
            <CopyIcon />
            Copy Connection Secret
          </Button>
          {/* For now not showing open in-app, only works well on Android, not on Desktop or iOS */}
          {/* <ExternalLinkButton to={pairingUri} variant="outline">
            <ExternalLinkIcon />
            Open In App
          </ExternalLinkButton> */}
        </div>
      </CardContent>
    </Card>
  );
}
