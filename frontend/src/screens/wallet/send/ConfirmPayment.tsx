import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";

import type { Invoice } from "@getalby/lightning-tools/bolt11";
import { ArrowLeftIcon } from "lucide-react";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { PaymentFailedAlert } from "src/components/PaymentFailedAlert";
import { PendingPaymentAlert } from "src/components/PendingPaymentAlert";
import { SpendingAlert } from "src/components/SpendingAlert";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useBalances } from "src/hooks/useBalances";
import { PayInvoiceResponse, TransactionMetadata } from "src/types";
import { request } from "src/utils/request";

export default function ConfirmPayment() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { data: balances } = useBalances();

  const invoice = state?.args?.paymentRequest as Invoice;
  const metadata = state?.args?.metadata as TransactionMetadata;
  const [isLoading, setLoading] = React.useState(false);
  const [errorMessage, setErrorMessage] = React.useState("");

  const confirmPayment = async () => {
    setErrorMessage("");
    try {
      setLoading(true);
      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${invoice.paymentRequest}`,
        {
          method: "POST",
          body: JSON.stringify({
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
          pageTitle: "Pay Invoice",
          invoice,
        },
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
      <AppHeader title="Pay Invoice" />
      <div className="max-w-lg grid gap-4">
        <PendingPaymentAlert />
        {errorMessage && (
          <PaymentFailedAlert
            errorMessage={errorMessage}
            invoice={invoice.paymentRequest}
          />
        )}
      </div>
      <div className="w-full md:max-w-lg">
        <Card>
          <CardHeader>
            <CardTitle className="text-center">Confirm Payment</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-6 pt-2">
            <div className="flex flex-col gap-1 items-center">
              <p className="text-2xl font-medium slashed-zero">
                {new Intl.NumberFormat().format(invoice.satoshi)} sats
              </p>
              <FormattedFiatAmount
                amount={invoice.satoshi}
                className="text-xl"
              />
            </div>
            {invoice.description && (
              <p className="text-lg text-muted-foreground break-anywhere">
                {invoice.description}
              </p>
            )}
          </CardContent>
          <CardFooter className="flex flex-col gap-2 pt-2">
            <SpendingAlert className="mb-2" amount={invoice.satoshi} />
            <LoadingButton
              onClick={confirmPayment}
              loading={isLoading}
              type="submit"
              className="w-full"
              autoFocus
            >
              Confirm Payment
            </LoadingButton>
            <div className="flex items-center justify-between gap-2 text-muted-foreground text-xs sensitive slashed-zero">
              Spending Balance:{" "}
              {new Intl.NumberFormat().format(
                Math.floor(balances.lightning.totalSpendable / 1000)
              )}{" "}
              sats
            </div>
            <LinkButton to="/wallet/send" variant="link" className="w-full">
              <ArrowLeftIcon className="w-4 h-4 mr-2" />
              Back
            </LinkButton>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
