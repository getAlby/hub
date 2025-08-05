import { Invoice } from "@getalby/lightning-tools/bolt11";
import React, { useEffect } from "react";
import AppHeader from "src/components/AppHeader";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { useToast } from "src/components/ui/use-toast";
import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

export function Bitrefill() {
  const { toast } = useToast();
  const [paymentDialogOpen, setPaymentDialogOpen] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [invoice, setInvoice] = React.useState<Invoice>();

  useEffect(() => {
    const handleIframeMessage = (event: MessageEvent) => {
      if (event.origin === "https://embed.bitrefill.com" && event.data) {
        // Some of the events have objects in event.data 🤷‍♂️
        try {
          const parsedData = JSON.parse(event.data);

          if (parsedData.event === "payment_intent") {
            setPaymentDialogOpen(true);
            const invoice = new Invoice({ pr: parsedData.paymentAddress });
            setInvoice(invoice);
          }
        } catch (e) {
          /* empty */
        }
      }
    };

    window.addEventListener("message", handleIframeMessage);

    return () => {
      window.removeEventListener("message", handleIframeMessage);
    };
  }, []);

  async function confirmPayment() {
    if (!invoice) {
      return;
    }

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

      if (!payInvoiceResponse?.preimage) {
        throw new Error("No preimage in response");
      }

      setPaymentDialogOpen(false);
      setInvoice(undefined);

      toast({
        title: "Payment successful",
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
  }

  return (
    <>
      <div className="flex flex-col gap-5 h-full">
        <AppHeader
          title="Bitrefill"
          description="Live on bitcoin by purchasing digital gift cards, eSIMs, and phone refills"
        />
        <iframe
          width="100%"
          className="grow rounded-xl"
          src={`https://embed.bitrefill.com/?ref=V6DBkUhx&showPaymentInfo=false&paymentMethod=lightning&showPaymentInfo=false&utm_source=alby_hub`}
          sandbox="allow-same-origin allow-popups allow-scripts allow-forms"
        ></iframe>
      </div>
      <AlertDialog open={paymentDialogOpen} onOpenChange={setPaymentDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm payment</AlertDialogTitle>
          </AlertDialogHeader>
          <div className="flex flex-col gap-3">
            <div>
              <div className="font-medium">Description</div>
              <div>{invoice?.description}</div>
            </div>
            <div>
              <div className="font-medium">Amount</div>
              <div className="flex flex-row gap-2 items-center">
                <span className="font-medium slashed-zero">
                  {new Intl.NumberFormat().format(invoice?.satoshi || 0)} sats
                </span>
                <FormattedFiatAmount
                  className="text-muted-foreground"
                  amount={invoice?.satoshi || 0}
                />
              </div>
            </div>
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction disabled={loading} onClick={confirmPayment}>
              {loading && <Loading className="w-4 h-4 mr-2" />}
              Pay now
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
