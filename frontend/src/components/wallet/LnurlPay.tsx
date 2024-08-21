import { LightningAddress } from "@getalby/lightning-tools";
import React from "react";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import PaymentSuccessCard from "src/components/wallet/PaymentSuccessCard";

import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

type LnurlPayProps = {
  lnAddress: LightningAddress;
  onReset: () => void;
  onAmountChange: React.Dispatch<React.SetStateAction<number>>;
};

function LnurlPay({ lnAddress, onReset, onAmountChange }: LnurlPayProps) {
  const [amount, setAmount] = React.useState("");
  const [comment, setComment] = React.useState("");
  const [preimage, setPreimage] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const { toast } = useToast();

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
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
        setPreimage(payInvoiceResponse.preimage);
        toast({
          title: "Successfully paid invoice",
        });
      }
      setAmount("");
      setComment("");
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

  return (
    <>
      {!preimage ? (
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
                  onAmountChange(parseInt(e.target.value));
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
            <LoadingButton loading={isLoading} type="submit" disabled={!amount}>
              Confirm Payment
            </LoadingButton>
            <Button onClick={onReset} variant="secondary">
              Back
            </Button>
          </div>
        </form>
      ) : (
        <PaymentSuccessCard preimage={preimage} onReset={onReset} />
      )}
    </>
  );
}

export default LnurlPay;
