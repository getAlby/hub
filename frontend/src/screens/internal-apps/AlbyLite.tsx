import { Invoice, LightningAddress } from "@getalby/lightning-tools";
import { AlertCircleIcon, CopyIcon, PartyPopperIcon } from "lucide-react";
import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest, UpdateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

// TODO: Check beforehand if username is available
export function AlbyLite() {
  const { toast } = useToast();
  const { data: apps } = useApps();
  const [username, setUsername] = React.useState("");
  const [nip05Pubkey, setNip05Pubkey] = React.useState("");
  const [invoice, setInvoice] = React.useState<Invoice | null>(null);
  const [lnAddress, setLnAddress] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const name = React.useMemo(() => `${username}@lite.albylabs.com`, [username]);

  React.useEffect(() => {
    (async () => {
      const ln = new LightningAddress("hello@getalby.com");
      await ln.fetch();
      if (!ln.lnurlpData) {
        throw new Error("invalid recipient lightning address");
      }
      const invoice = await ln.requestInvoice({
        satoshi: 21000,
        comment: "Invoice for alby lite address",
      });
      setInvoice(invoice);
    })();
  }, []);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);

    try {
      const isPaid = await invoice?.isPaid();
      if (!isPaid) {
        throw new Error("Please pay the invoice to proceed.");
      }

      if (apps?.some((existingApp) => existingApp.name === name)) {
        throw new Error("A connection with the same name already exists.");
      }

      const createAppRequest: CreateAppRequest = {
        name,
        scopes: ["make_invoice", "notifications"],
        budgetRenewal: "never",
        maxAmount: 0,
        isolated: true,
        metadata: {
          app_store_app_id: "alby-lite",
        },
      };

      const createAppResponse = await createApp(createAppRequest);

      // TODO: proxy through hub backend and remove CSRF exceptions for lite.albylabs.com
      const createLNAddressResponse = await fetch(
        "https://lite.albylabs.com/users",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            username,
            nostrPubkey: nip05Pubkey,
            connectionSecret: createAppResponse.pairingUri,
          }),
        }
      );
      if (!createLNAddressResponse.ok) {
        throw new Error(
          "Failed to create ln address: " + createLNAddressResponse.status
        );
      }

      const { lightningAddress } = await createLNAddressResponse.json();
      if (!lightningAddress) {
        throw new Error("No lightning address in response");
      }

      // add the lightning address to the app metadata
      const updateAppRequest: UpdateAppRequest = {
        name: createAppRequest.name,
        scopes: createAppRequest.scopes,
        expiresAt: createAppRequest.expiresAt,
        budgetRenewal: createAppRequest.budgetRenewal!,
        isolated: createAppRequest.isolated!,
        maxAmount: createAppRequest.maxAmount!,
        metadata: {
          ...createAppRequest.metadata,
          lightning_address: lightningAddress,
        },
      };

      await request(`/api/apps/${createAppResponse.pairingPublicKey}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateAppRequest),
      });
      setLnAddress(lightningAddress);
      toast({ title: "🎉 Lightning address created" });
    } catch (error) {
      handleRequestError(toast, "Failed to create lightning address", error);
    }
    setLoading(false);
  };

  const onboardedApps = apps?.filter(
    (app) => app.metadata?.app_store_app_id === "alby-lite"
  );

  const copy = () => {
    copyToClipboard(invoice?.paymentRequest || "", toast);
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Alby Lite"
        description="Create lightning addresses for your app connections"
      />
      {!lnAddress && (
        <>
          {!!onboardedApps?.length && (
            <>
              <form
                onSubmit={handleSubmit}
                className="flex flex-col items-start gap-5 max-w-lg"
              >
                <div className="w-full grid gap-1.5">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    autoFocus
                    type="text"
                    name="username"
                    value={username}
                    id="username"
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    autoComplete="off"
                    placeholder="Choose your desired username"
                  />
                </div>
                <div className="w-full grid gap-1.5">
                  <Label htmlFor="nip05">NIP-05 Pubkey (npub or hex)</Label>
                  <Input
                    type="text"
                    name="nip05"
                    value={nip05Pubkey}
                    id="nip05"
                    onChange={(e) => setNip05Pubkey(e.target.value)}
                    required
                    autoComplete="off"
                    placeholder="Enter your NIP-05 pubkey"
                  />
                </div>
                {invoice ? (
                  <div>
                    <p className="text-sm font-medium">
                      Please pay the invoice
                    </p>
                    <div className="flex flex-col gap-2">
                      <QRCode value={invoice.paymentRequest} />
                      <Button onClick={copy} variant="outline" type="button">
                        <CopyIcon className="w-4 h-4 mr-2" />
                        Copy Invoice
                      </Button>
                    </div>
                  </div>
                ) : (
                  <Loading />
                )}
                <div className="flex items-center gap-2">
                  <AlertCircleIcon className="h-4 w-4" />
                  <p className="text-sm">
                    Proceeding would create an app connection for the lightning
                    address
                  </p>
                </div>
                <LoadingButton loading={isLoading} type="submit">
                  Create Lightning Address
                </LoadingButton>
              </form>
              <p className="text-sm text-muted-foreground">
                You've created {onboardedApps.length} lightning addresses so
                far.
              </p>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
                {onboardedApps.map((app, index) => (
                  <AppCard key={index} app={app} />
                ))}
              </div>{" "}
            </>
          )}
        </>
      )}
      {lnAddress && (
        <div className="max-w-lg flex flex-col gap-5">
          <Alert>
            <PartyPopperIcon className="h-4 w-4" />
            <AlertTitle>Your lightning address is ready!</AlertTitle>
            <AlertDescription>
              You can use it to receive payments, set up your NIP-05
              verification, and even accept zaps on Nostr.
            </AlertDescription>
          </Alert>
          <div className="flex flex-col items-center relative">
            <QRCode value={lnAddress} />
          </div>
          <div className="flex gap-2">
            <Input disabled readOnly type="text" value={lnAddress} />
            <Button
              onClick={() => copyToClipboard(lnAddress, toast)}
              variant="outline"
            >
              <CopyIcon className="w-4 h-4 mr-2" />
              Copy
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
