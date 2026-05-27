import React from "react";
import { Label } from "src/components/ui/label";

import type { Invoice } from "@getalby/lightning-tools/bolt11";
import { XIcon } from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { CurrencyInputField } from "src/components/CurrencyInputField";
import { InsufficientLightningBalanceAlert } from "src/components/InsufficientLightningBalanceAlert";
import Loading from "src/components/Loading";
import { PaymentFailedAlert } from "src/components/PaymentFailedAlert";
import { PendingPaymentAlert } from "src/components/PendingPaymentAlert";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useBalances } from "src/hooks/useBalances";
import PayFromSelect from "src/screens/wallet/send/PayFromSelect";
import { PayInvoiceRequest, PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

export default function ZeroAmount() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { data: balances } = useBalances();

  const invoice = state?.args?.paymentRequest as Invoice;
  const [appId, setAppId] = React.useState<number>();
  const [amountSat, setAmountSat] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const [errorMessage, setErrorMessage] = React.useState("");

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage("");
    try {
      if (!invoice) {
        throw new Error("no invoice set");
      }
      setLoading(true);
      const payload: PayInvoiceRequest = {
        amountMsat: +amountSat * 1000,
        fromAppId: appId,
      };
      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${invoice.paymentRequest}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(payload),
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
          amountSat,
        },
        replace: true,
      });
      toast("Successfully paid invoice");
    } catch (e) {
      console.error(e);
      setErrorMessage("" + e);
      toast.error("Failed to send payment", {
        description: "" + e,
      });
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
      <AppHeader pageTitle="Pay Invoice" title="Pay Invoice" />
      <div className="max-w-lg grid gap-4">
        <PendingPaymentAlert />
        {errorMessage && (
          <PaymentFailedAlert
            errorMessage={errorMessage}
            invoice={invoice.paymentRequest}
          />
        )}
      </div>
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
        <CurrencyInputField
          id="amount"
          valueSat={amountSat}
          onValueSatChange={setAmountSat}
          minSat={1}
          maxSat={balances.lightning.totalSpendableSat}
          required
          autoFocus
          contextRows={[
            {
              label: "Lightning balance",
              amountSat: balances.lightning.totalSpendableSat,
            },
          ]}
        />
        <PayFromSelect appId={appId} onChange={setAppId} />
        <InsufficientLightningBalanceAlert amountSat={+amountSat} />
        <div className="flex gap-2">
          <LinkButton to="/wallet/send" variant="outline">
            Back
          </LinkButton>
          <LoadingButton loading={isLoading} type="submit" className="flex-1">
            Send
          </LoadingButton>
        </div>
      </form>
    </div>
  );
}
