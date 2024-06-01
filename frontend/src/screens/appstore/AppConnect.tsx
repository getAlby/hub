import { CopyIcon } from "lucide-react";
import { useEffect, useState } from "react";
import {
  Link,
  Navigate,
  useLocation,
  useNavigate,
  useParams,
} from "react-router-dom";
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

export default function AppConnect() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();
  const params = useParams();
  const [timeout, setTimeout] = useState(false);
  const createAppResponse = state as CreateAppResponse;
  const appstoreApp = suggestedApps.find((x) => x.id == params.id);
  const { data: app } = useApp(createAppResponse.pairingPublicKey, true);
  const pairingUri = createAppResponse.pairingUri;

  const copy = () => {
    copyToClipboard(pairingUri);
    toast({ title: "Copied to clipboard." });
  };

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      setTimeout(true);
    }, 10000);

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

  if (!createAppResponse || !appstoreApp) {
    return <Navigate to="/apps/new" />;
  }

  return (
    <>
      <AppHeader
        title={`Connect to ${appstoreApp.title}`}
        description="Configure wallet permissions for the app and follow instructions to finalise the connection"
      />
      <div className="flex flex-col gap-3 ph-no-capture">
        <div>
          <p>
            1. Open{" "}
            <ExternalLink
              className="font-semibold underline"
              to={appstoreApp.to}
            >
              {appstoreApp.title}
            </ExternalLink>{" "}
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
              <QRCode value={pairingUri} className="w-full" />
              <img
                src={appstoreApp.logo}
                className="absolute w-12 h-12 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-muted p-1 rounded-xl"
              />
            </a>
            <div>
              <Button onClick={copy} variant="outline">
                <CopyIcon className="w-4 h-4 mr-2" />
                Copy
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
