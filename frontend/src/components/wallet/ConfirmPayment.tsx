import { Invoice } from "@getalby/lightning-tools";
import React from "react";
import { Button } from "src/components/ui/button";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import PaymentSuccessCard from "src/components/wallet/PaymentSuccessCard";

import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

type ConfirmPaymentProps = {
  invoice: Invoice;
  onReset: () => void;
};

function ConfirmPayment({ invoice, onReset }: ConfirmPaymentProps) {
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);
  const [preimage, setPreimage] = React.useState("");

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
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
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to send: " + e,
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
            <p className="font-medium text-lg mb-2">Payment Details</p>
            <div>
              <Label>Amount</Label>
              <p className="font-bold">{invoice.satoshi} sats</p>
            </div>
            {invoice.description && (
              <div className="mt-2">
                <Label>Description</Label>
                <p className="text-muted-foreground">{invoice.description}</p>
              </div>
            )}
          </div>
          <div className="flex gap-4">
            <LoadingButton loading={isLoading} type="submit">
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

export default ConfirmPayment;
