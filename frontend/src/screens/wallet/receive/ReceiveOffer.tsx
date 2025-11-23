import { CircleAlertIcon, CopyIcon, RefreshCwIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import QRCode from "src/components/QRCode";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";

import { copyToClipboard } from "src/lib/clipboard";
import { CreateOfferRequest } from "src/types";
import { request } from "src/utils/request";

export default function ReceiveOffer() {
  const [isLoading, setLoading] = React.useState(false);
  const [description, setDescription] = React.useState<string>("");
  const [offer, setOffer] = React.useState<string | null>(null);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    try {
      setLoading(true);
      const offer = await request<string>("/api/offers", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          description,
        } as CreateOfferRequest),
      });

      if (offer) {
        setOffer(offer);

        toast("Successfully created BOLT-12 offer");
      }
    } catch (e) {
      toast.error("Failed to create offer", {
        description: "" + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const copy = () => {
    copyToClipboard(offer as string);
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title={offer ? "Lightning Offer" : "Create Lightning Offer"}
        description={
          !offer ? "Create a reusable, non-expiring lightning invoice" : ""
        }
      />
      <div className="grid gap-6 md:max-w-lg">
        {!offer && (
          <Alert>
            <CircleAlertIcon className="h-4 w-4" />
            <AlertTitle>BOLT-12 Offers are in beta</AlertTitle>
            <AlertDescription>
              BOLT-12 is not supported by all wallets and nodes in the lightning
              network. This feature will work only if you have a channel with a
              node that supports onion message forwarding, and are paid by a
              lightning wallet that supports paying BOLT-12 offers.
            </AlertDescription>
          </Alert>
        )}
        {offer ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-center">Lightning Offer</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col items-center gap-6">
              <QRCode value={offer} className="w-full" />
              {description && (
                <p className="text-muted-foreground my-2">{description}</p>
              )}
            </CardContent>
            <CardFooter className="flex flex-col gap-2">
              <Button className="w-full" onClick={copy} variant="secondary">
                <CopyIcon className="w-4 h-4 mr-2" />
                Copy Offer
              </Button>
              <Button
                className="w-full"
                variant="outline"
                onClick={() => {
                  setDescription("");
                  setOffer(null);
                }}
              >
                <RefreshCwIcon className="h-4 w-4 shrink-0 mr-2" />
                New Offer
              </Button>
            </CardFooter>
          </Card>
        ) : (
          <form onSubmit={handleSubmit} className="grid gap-6">
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                type="text"
                value={description}
                placeholder="For e.g. what is this payment for?"
                onChange={(e) => {
                  setDescription(e.target.value);
                }}
              />
            </div>
            <div>
              <LoadingButton
                className="w-full md:w-auto"
                loading={isLoading}
                type="submit"
              >
                Create Offer
              </LoadingButton>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
