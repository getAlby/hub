import { CopyIcon, EyeIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import QRCode from "src/components/QRCode";
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
  showToasts = false,
}: {
  app: App;
  pairingUri: string;
  appStoreApp?: AppStoreApp;
  showToasts?: boolean;
}) {
  const navigate = useNavigate();
  const [isQRCodeVisible, setIsQRCodeVisible] = useState(false);
  const effectRan = useRef(false);

  const copy = () => {
    copyToClipboard(pairingUri);
  };

  useEffect(() => {
    if (!showToasts) {
      return;
    }

    if (!app.lastUsedAt && !effectRan.current) {
      effectRan.current = true;

      const toastId = toast.loading("Waiting for app to connect", {
        description: "Scan the QR code or copy the connection secret",
      });

      const timeoutId = window.setTimeout(() => {
        toast.dismiss(toastId);
        const timeoutToastId = toast.loading(
          "Connection taking longer than usual",
          {
            action: {
              label: "Continue anyway",
              onClick: () => {
                toast.dismiss(timeoutToastId);
                navigate(`/apps/${app?.id}`);
              },
            },
          }
        );
      }, 30000);

      return () => {
        window.clearTimeout(timeoutId);
        toast.dismiss(toastId);
      };
    } else if (app.lastUsedAt) {
      toast.dismiss(); // Clear any existing toasts
      toast.success("App connected", {
        description: "Your app is now connected to the sub-wallet",
      });

      setTimeout(() => {
        navigate(`/apps/${app.id}`);
      }, 3000);
    }
  }, [app.lastUsedAt, app.id, showToasts, navigate]);

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-center">Scan To Connect</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-5">
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
