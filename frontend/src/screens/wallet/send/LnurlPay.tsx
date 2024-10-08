import React from "react";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { LightningAddress } from "@getalby/lightning-tools";
import { Link, useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";

export default function LnurlPay() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();

  const lnAddress = state?.args?.lnAddress as LightningAddress;
  const [amount, setAmount] = React.useState("");
  const [comment, setComment] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (!lnAddress) {
        throw new Error("no lightning address set");
      }
      setLoading(true);
      const invoice = await lnAddress.requestInvoice({
        satoshi: parseInt(amount),
        comment,
      });
      navigate(`/wallet/send/confirm-payment`, {
        state: {
          args: { paymentRequest: invoice },
        },
      });
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to send payment",
        description: "" + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    if (!lnAddress) {
      navigate("/wallet/send");
    }
  }, [navigate, lnAddress]);

  if (!lnAddress) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-4">
      <div>
        <p className="font-medium text-lg mb-2">{lnAddress.address}</p>
        {lnAddress.lnurlpData?.description && (
          <div className="mb-2">
            <Label>Description</Label>
            <p className="text-muted-foreground">
              {lnAddress.lnurlpData.description}
            </p>
          </div>
        )}
        <div className="mb-2">
          <Label htmlFor="amount">Amount</Label>
          <Input
            id="amount"
            type="number"
            value={amount}
            placeholder="Amount in Satoshi..."
            onChange={(e) => {
              setAmount(e.target.value.trim());
            }}
            min={1}
            required
            autoFocus
          />
        </div>
        {!!lnAddress.lnurlpData?.commentAllowed && (
          <div className="mb-2">
            <Label htmlFor="comment">Comment</Label>
            <Input
              id="comment"
              type="text"
              value={comment}
              placeholder="Optional"
              onChange={(e) => {
                setComment(e.target.value);
              }}
            />
          </div>
        )}
      </div>
      <div className="flex gap-4">
        <LoadingButton loading={isLoading} type="submit">
          Continue
        </LoadingButton>
        <Link to="/wallet/send">
          <Button variant="secondary">Back</Button>
        </Link>
      </div>
    </form>
  );
}
