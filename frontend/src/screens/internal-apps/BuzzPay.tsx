import { CopyIcon } from "lucide-react";
import { ExternalLinkIcon } from "lucide-react";
import buzzpay from "src/assets/suggested-apps/buzzpay.png";
import React from "react";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import QRCode from "src/components/QRCode";
import { LoadingButton } from "src/components/ui/loading-button";
import { copyToClipboard } from "src/lib/clipboard";
import { openLink } from "src/utils/openLink";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function BuzzPay() {
  const [name, setName] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const { data: apps, mutate: reloadApps } = useApps();
  const [posUrl, setPosUrl] = React.useState("");
  const { toast } = useToast();

  if (!apps) {
    return <Loading />;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        if (apps?.some((existingApp) => existingApp.name === name)) {
          throw new Error("A connection with the same name already exists.");
        }

        const createAppResponse = await createApp({
          name,
          scopes: ["get_info", "lookup_invoice", "make_invoice"],
          isolated: true,
          metadata: {
            app_store_app_id: "buzzpay",
          },
        });

        setPosUrl(
          `https://pos.albylabs.com/?name=${encodeURIComponent(name)}#/wallet/${encodeURIComponent(createAppResponse.pairingUri)}/new`
        );

        await reloadApps();

        toast({ title: "BuzzPay PoS connection created" });
      } catch (error) {
        handleRequestError(toast, "Failed to create PoS connection", error);
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
            <strong>🎉 PoS connection created!</strong>
            <br />
            Open your PoS link below and share it with the employees and devices
            you want to use.
            <br />
            Please save this link. It will not be shown again but you can create
            new PoS connections for your Alby Hub anytime.
          </p>

          <div className="flex gap-2">
            <Input disabled readOnly type="text" value={posUrl} />
            <Button
              onClick={() => copyToClipboard(posUrl, toast)}
              variant="outline"
            >
              <CopyIcon className="w-4 h-4 mr-2" />
              Copy
            </Button>
            <Button onClick={() => openLink(posUrl)} variant="outline">
              <ExternalLinkIcon className="w-4 h-4 mr-2" />
              Open
            </Button>
          </div>
          <div>
            <QRCode className={"w-full"} value={posUrl} />
            <img
              src={buzzpay}
              className="absolute w-12 h-12 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-muted p-1 rounded-xl"
            />
          </div>
        </div>
      )}
      {!posUrl && (
        <div className="max-w-lg flex flex-col gap-5">
          <p className="text-muted-foreground">
            BuzzPay works by creating read-only connections to your Alby Hub.
            You can share the link safely with employees and use it on any
            device.
          </p>
          <ul className="text-muted-foreground">
            <li>🔒 One-Click Setup: Secure read-only link</li>
            <li>🔗 Sharable with all your employees and devices</li>
            <li>⚡ Lightning-Fast Transactions directly to your Alby Hub</li>
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
      )}
    </div>
  );
}
