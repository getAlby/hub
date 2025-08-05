import { CircleAlertIcon, CopyIcon, RefreshCwIcon } from "lucide-react";
import React from "react";
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
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { copyToClipboard } from "src/lib/clipboard";
import { CreateOfferRequest } from "src/types";
import { request } from "src/utils/request";

export default function ReceiveOffer() {
  const { toast } = useToast();
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

        toast({
          title: "Successfully created BOLT-12 offer",
        });
      }
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to create offer: " + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const copy = () => {
    copyToClipboard(offer as string, toast);
  };

  return (
    <div className="grid gap-5 md:max-w-lg">
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
        <>
          <Card className="w-full md:max-w-xs">
            <CardHeader>
              <CardTitle className="text-center">Lightning Offer</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col items-center gap-4">
              <QRCode value={offer} className="w-full" />
              <div className="flex flex-col md:flex-row gap-4 w-full">
                <Button
                  className="flex-1"
                  variant="outline"
                  onClick={() => {
                    setDescription("");
                    setOffer(null);
                  }}
                >
                  <RefreshCwIcon className="h-4 w-4 shrink-0 mr-2" />
                  New Offer
                </Button>
                <Button className="flex-1" onClick={copy} variant="secondary">
                  <CopyIcon />
                  Copy
                </Button>
              </div>
            </CardContent>
          </Card>
        </>
      ) : (
        <form onSubmit={handleSubmit} className="grid gap-5">
          <div>
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
  );
}
