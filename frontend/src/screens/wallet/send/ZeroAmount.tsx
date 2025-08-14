import React from "react";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";

import type { Invoice } from "@getalby/lightning-tools/bolt11";
import { XIcon } from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { PendingPaymentAlert } from "src/components/PendingPaymentAlert";
import { SpendingAlert } from "src/components/SpendingAlert";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useBalances } from "src/hooks/useBalances";
import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

export default function ZeroAmount() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { data: balances } = useBalances();

  const invoice = state?.args?.paymentRequest as Invoice;
  const [amount, setAmount] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (!invoice) {
        throw new Error("no invoice set");
      }
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
      if (!payInvoiceResponse?.preimage) {
        throw new Error("No preimage in response");
      }
      navigate(`/wallet/send/success`, {
        state: {
          preimage: payInvoiceResponse.preimage,
          pageTitle: "Pay Invoice",
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

  return (
    <div className="grid gap-4">
      <AppHeader title="Pay Invoice" />
      <PendingPaymentAlert />
      <form onSubmit={onSubmit} className="grid gap-6 md:max-w-lg">
        <div className="grid gap-2">
          <div className="text-sm font-medium">Recipient</div>
          <div className="flex items-center justify-between gap-2">
            <p className="text-sm break-all line-clamp-1">
              {invoice.paymentRequest}
            </p>
            <Link to="/wallet/send">
              <XIcon className="w-4 h-4 cursor-pointer text-muted-foreground" />
            </Link>
          </div>
        </div>
        {invoice.description && (
          <div className="grid gap-2">
            <Label>Description</Label>
            <p className="text-muted-foreground text-sm truncate max-w-full">
              {invoice.description}
            </p>
          </div>
        )}
        <div className="grid gap-2">
          <Label htmlFor="amount">Amount</Label>
          <InputWithAdornment
            id="amount"
            type="number"
            value={amount}
            placeholder="Amount in Satoshi..."
            onChange={(e) => {
              setAmount(e.target.value.trim());
            }}
            min={1}
            max={Math.floor(balances.lightning.totalSpendable / 1000)}
            required
            autoFocus
            className="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
            endAdornment={
              <FormattedFiatAmount amount={Number(amount)} className="mr-2" />
            }
          />
          <div className="grid gap-2">
            <div className="flex justify-between text-xs text-muted-foreground sensitive slashed-zero">
              <div>
                Spending Balance:{" "}
                {new Intl.NumberFormat().format(
                  Math.floor(balances.lightning.totalSpendable / 1000)
                )}{" "}
                sats
              </div>
              <FormattedFiatAmount
                className="text-xs"
                amount={balances.lightning.totalSpendable / 1000}
              />
            </div>
          </div>
        </div>
        <SpendingAlert amount={+amount} />
        <div className="flex gap-2">
          <LinkButton to="/wallet/send" variant="outline">
            Back
          </LinkButton>
          <LoadingButton
            loading={isLoading}
            type="submit"
            className="w-full md:w-fit"
          >
            Send
          </LoadingButton>
        </div>
      </form>
    </div>
  );
}
