import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";

import type { Invoice } from "@getalby/lightning-tools/bolt11";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { SpendingAlert } from "src/components/SpendingAlert";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { PayInvoiceResponse, TransactionMetadata } from "src/types";
import { request } from "src/utils/request";

export default function ConfirmPayment() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();

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
          invoice,
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

  if (!balances || !invoice) {
    return <Loading />;
  }

  const maxSpendable = Math.max(
    balances.lightning.nextMaxSpendableMPP -
      Math.max(
        0.01 * balances.lightning.nextMaxSpendableMPP,
        10000 /* fee reserve */
      ),
    0
  );

  return (
    <form onSubmit={onSubmit} className="grid gap-4">
      <div className="grid gap-2">
        <p className="font-medium text-lg">Payment Details</p>
        <div>
          <Label>Amount</Label>
          <p className="text-xl font-bold slashed-zero">
            {new Intl.NumberFormat().format(amount || invoice.satoshi)} sats
          </p>
          <FormattedFiatAmount amount={amount || invoice.satoshi} />
        </div>
        {invoice.description && (
          <div className="break-all">
            <Label>Description</Label>
            <p className="text-muted-foreground">{invoice.description}</p>
          </div>
        )}
      </div>
      {hasChannelManagement &&
        (amount || invoice.satoshi || 0) * 1000 >= maxSpendable && (
          <SpendingAlert maxSpendable={maxSpendable} />
        )}
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
