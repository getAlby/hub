import { DialogDescription, DialogTrigger } from "@radix-ui/react-dialog";
import { CopyIcon, ScanIcon } from "lucide-react";
import { useEffect } from "react";
import { Link, Navigate, useLocation } from "react-router-dom";

import QRCode from "src/components/QRCode";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle
} from "src/components/ui/dialog";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";
import { CreateAppResponse } from "src/types";

export default function AppCreated() {
  const { state } = useLocation();
  const { toast } = useToast();
  const createAppResponse = state as CreateAppResponse;

  useEffect(() => {
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
  }, []);

  if (!createAppResponse) {
    return <Navigate to="/apps/new" />;
  }

  const pairingUri = createAppResponse.pairingUri;

  const copy = () => {
    copyToClipboard(pairingUri);
    toast({ title: "Copied to clipboard." });
  };

  return (
    <div className="w-full max-w-screen-sm mx-auto mt-6 md:px-4 ph-no-capture">
      <h2 className="font-bold text-2xl font-headline mb-2 text-center">
        🚀 Almost there!
      </h2>
      <div className="font-medium text-center mb-6">
        Complete the last step of the setup by pasting or scanning your
        connection's pairing secret in the desired app to finalise the
        connection.
      </div>

      <div className="flex flex-col items-center">
        <Link to={pairingUri}>
          <Button size="lg">
            <NostrWalletConnectIcon className="w-6 h-6 mr-2" />
            Open in supported app
          </Button>
        </Link>
        <div className="text-center text-xs text-muted-foreground mt-2">
          Only connect with apps you trust!
        </div>

        <div className="text-sm text-center mt-8 mb-1"></div>
        <div className="flex flex-col gap-3">
          <div className="text-center text-sm">Manually pair app ↓</div>
          <Button variant="secondary" onClick={copy}>
            <CopyIcon className="w-4 h-4 mr-2" />
            Copy pairing secret
          </Button>

          <Dialog>
            <DialogTrigger>
              <Button variant="secondary">
                <ScanIcon className="w-4 h-4 mr-2" />
                QR Code
              </Button>
            </DialogTrigger>

            <DialogContent>
              <DialogHeader>
                <DialogTitle>
                  Scan QR Code
                </DialogTitle>
                <DialogDescription>
                  Open the app you want to pair and scan this QR code to connect.
                </DialogDescription>
              </DialogHeader>
              <div className="flex flex-row justify-center p-3">
                <a
                  href={pairingUri}
                  target="_blank"
                >
                  <QRCode value={pairingUri} />
                </a>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>
    </div>
  );
}
