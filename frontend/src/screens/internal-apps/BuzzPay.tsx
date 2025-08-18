import { AlertTriangleIcon, CopyIcon, ExternalLinkIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import buzzpay from "src/assets/suggested-apps/buzzpay.png";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useApps } from "src/hooks/useApps";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";
import { openLink } from "src/utils/openLink";

export function BuzzPay() {
  const [name, setName] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const { data: appsData, mutate: reloadApps } = useApps(undefined, undefined, {
    appStoreAppId: "buzzpay",
  });
  const [posUrl, setPosUrl] = React.useState("");

  if (!appsData) {
    return <Loading />;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        const createAppResponse = await createApp({
          name,
          scopes: ["get_info", "lookup_invoice", "make_invoice"],
          isolated: true,
          metadata: {
            app_store_app_id: "buzzpay",
          },
        });

        setPosUrl(
          `https://pos.albylabs.com/#/?nwc=${btoa(createAppResponse.pairingUri)}&label=${encodeURIComponent(name)}`
        );

        await reloadApps();

        toast("BuzzPay PoS connection created");
      } catch (error) {
        handleRequestError("Failed to create PoS connection", error);
      }
      setLoading(false);
    })();
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="BuzzPay"
        description="The easiest Bitcoin Point-of-Sale (PoS) system."
      />
      {posUrl && (
        <div className="max-w-lg flex flex-col gap-5">
          <p>
            Open the PoS link below and share it with your employees and devices
            you want to use the PoS with.
          </p>
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

          <div className="flex flex-col items-center relative">
            <QRCode value={posUrl} />
            <img
              src={buzzpay}
              className="absolute w-12 h-12 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-muted p-1 rounded-xl"
            />
          </div>
          <div className="flex gap-2">
            <Input disabled readOnly type="text" value={posUrl} />
            <Button onClick={() => copyToClipboard(posUrl)} variant="outline">
              <CopyIcon />
              Copy
            </Button>
            <Button onClick={() => openLink(posUrl)} variant="outline">
              <ExternalLinkIcon />
              Open
            </Button>
          </div>
        </div>
      )}
      {!posUrl && (
        <>
          <div className="max-w-lg flex flex-col gap-5">
            <p className="text-muted-foreground">
              BuzzPay works by creating read-only connections to your Alby Hub.
            </p>
            <ul className="text-muted-foreground">
              <li>🔒 Allow employees to collect but never spend your funds</li>
              <li>🔗 Sharable link that can be used on any device</li>
              <li>⚡ Lightning fast transactions directly to your Alby Hub</li>
            </ul>
            <form
              onSubmit={handleSubmit}
              className="flex flex-col items-start gap-5 max-w-lg"
            >
              <div className="w-full grid gap-1.5">
                <Label htmlFor="name">PoS Name</Label>
                <Input
                  autoFocus
                  type="text"
                  name="name"
                  value={name}
                  id="name"
                  onChange={(e) => setName(e.target.value)}
                  required
                  autoComplete="off"
                />
              </div>
              <LoadingButton loading={isLoading} type="submit">
                Next
              </LoadingButton>
            </form>
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
            {appsData.apps.map((app, index) => (
              <AppCard key={index} app={app} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
