import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { Button } from "src/components/ui/button";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { Invoice } from "@getalby/lightning-tools";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { PayInvoiceResponse, TransactionMetadata } from "src/types";
import { request } from "src/utils/request";

export default function ConfirmPayment() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();

  const amount = state?.args?.amount as number | undefined;
  const invoice = state?.args?.paymentRequest as Invoice;
  const metadata = state?.args?.metadata as TransactionMetadata;
  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${invoice.paymentRequest}`,
        {
          method: "POST",
          body: JSON.stringify({
            amount: amount ? amount * 1000 : undefined,
            metadata,
          }),
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      if (!payInvoiceResponse?.preimage) {
        throw new Error("No preimage in response");
      }

      navigate(`/wallet/send/success`, {
        state: {
          preimage: payInvoiceResponse.preimage,
          amount: amount || invoice.satoshi,
        },
      });
      toast({
        title: "Successfully paid invoice",
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
    if (!invoice) {
      navigate("/wallet/send");
    }
  }, [navigate, invoice]);

  if (!invoice) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-4">
      <div>
        <p className="font-medium text-lg mb-2">Payment Details</p>
        <div>
          <Label>Amount</Label>
          <p className="text-xl font-bold slashed-zero">
            {new Intl.NumberFormat().format(amount || invoice.satoshi)} sats
          </p>
          <FormattedFiatAmount amount={amount || invoice.satoshi} />
        </div>
        {invoice.description && (
          <div className="mt-2 break-all">
            <Label>Description</Label>
            <p className="text-muted-foreground">{invoice.description}</p>
          </div>
        )}
      </div>
      <div className="flex gap-4">
        <LoadingButton loading={isLoading} type="submit" autoFocus>
          Confirm Payment
        </LoadingButton>
        <Link to="/wallet/send">
          <Button variant="secondary">Back</Button>
        </Link>
      </div>
    </form>
  );
}
