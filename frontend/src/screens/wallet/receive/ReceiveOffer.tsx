import { AlertTriangleIcon, CopyIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
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
    <div className="grid gap-5">
      <Alert>
        <AlertTriangleIcon className="h-4 w-4" />
        <AlertTitle>BOLT-12 is in beta phase</AlertTitle>
        <AlertDescription>
          This will only work if you have a channel with a node that supports
          onion message forwarding
        </AlertDescription>
      </Alert>
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          {offer ? (
            <>
              <Card className="w-full">
                <CardHeader>
                  <CardTitle className="text-center">BOLT-12 Offer</CardTitle>
                </CardHeader>
                <CardContent className="flex flex-col items-center gap-4">
                  <QRCode value={offer} className="w-full" />
                  <div>
                    <Button onClick={copy} variant="outline">
                      <CopyIcon className="w-4 h-4 mr-2" />
                      Copy Offer
                    </Button>
                  </div>
                </CardContent>
              </Card>
              <Button
                className="mt-4 w-full"
                onClick={() => {
                  setOffer(null);
                }}
              >
                Create Another Offer
              </Button>
              <Link to="/wallet">
                <Button className="mt-4 w-full" variant="secondary">
                  Back To Wallet
                </Button>
              </Link>
            </>
          ) : (
            <form onSubmit={handleSubmit} className="grid gap-5">
              <div>
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  type="text"
                  value={description}
                  placeholder="For e.g. OCEAN Payouts for bc1a..."
                  onChange={(e) => {
                    setDescription(e.target.value);
                  }}
                />
              </div>
              <div>
                <LoadingButton
                  loading={isLoading}
                  type="submit"
                  disabled={!description}
                >
                  Create Offer
                </LoadingButton>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
