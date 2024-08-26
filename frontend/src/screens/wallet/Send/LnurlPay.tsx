import React from "react";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { LightningAddress, LnUrlPayResponse } from "@getalby/lightning-tools";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

export default function LnurlPay() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();

  const lnurlDetails = state.args?.lnurlDetails as LnUrlPayResponse;
  const [amount, setAmount] = React.useState("");
  const [comment, setComment] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
      const lnAddress = new LightningAddress(lnurlDetails.identifier);
      // this is set instead of calling fetch because
      // requestInvoice uses lnurlpData.callback
      lnAddress.lnurlpData = lnurlDetails;
      const invoice = await lnAddress.requestInvoice({
        satoshi: parseInt(amount),
        comment,
      });
      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${invoice.paymentRequest}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      if (payInvoiceResponse) {
        navigate(`/wallet/send/success`, {
          state: {
            preimage: payInvoiceResponse.preimage,
          },
        });
        toast({
          title: "Successfully paid invoice",
        });
      }
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

  if (!state.args?.lnurlDetails) {
    navigate("/wallet/send");
    return null;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-4">
      <div>
        <p className="font-medium text-lg mb-2">{lnurlDetails.identifier}</p>
        {lnurlDetails?.description && (
          <div className="mb-2">
            <Label>Description</Label>
            <p className="text-muted-foreground">{lnurlDetails.description}</p>
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
        {!!lnurlDetails?.commentAllowed && (
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
        <LoadingButton loading={isLoading} type="submit" disabled={!amount}>
          Confirm Payment
        </LoadingButton>
        <Link to="/wallet/send">
          <Button variant="secondary">Back</Button>
        </Link>
      </div>
    </form>
  );
}
