import { Invoice, fiat } from "@getalby/lightning-tools";
import { CopyIcon, LightbulbIcon } from "lucide-react";
import React, { useState } from "react";
import { LightningIcon } from "src/components/icons/Lightning";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";
import useChannelOrderStore from "src/state/ChannelOrderStore";
import { LSPOrderResponse } from "src/types";
import { request } from "src/utils/request";

type PayLightningInvoiceProps = {
  invoice: string;
  lspOrderResponse?: LSPOrderResponse;
  canPayInternally?: boolean | undefined;
};

export function PayLightningInvoice({
  invoice,
  lspOrderResponse,
  canPayInternally,
}: PayLightningInvoiceProps) {
  const amount = new Invoice({
    pr: invoice,
  }).satoshi;
  const [fiatAmount, setFiatAmount] = React.useState(0);
  React.useEffect(() => {
    fiat
      .getFiatValue({ satoshi: amount, currency: "USD" })
      .then((fiatAmount) => setFiatAmount(fiatAmount));
  }, [amount]);
  const { toast } = useToast();
  const copy = () => {
    copyToClipboard(invoice, toast);
  };

  const [isPaying, setPaying] = useState(false);

  const handlePayment = async () => {
    try {
      setPaying(true);

      await request(`/api/payments/${lspOrderResponse?.invoice}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      useChannelOrderStore.getState().updateOrder({
        status: "paid",
      });

      toast({
        title: "Channel successfully requested",
      });
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to send: " + e,
      });
      console.error(e);
    } finally {
      setPaying(false);
    }
  };

  return (
    <div className="w-80 flex flex-col gap-6 px-8 py-6 items-center justify-center border rounded-xl">
      <div className="flex items-center justify-center gap-2">
        <Loading variant="loader" />
        <p className="text-secondary-foreground">Waiting for payment...</p>
      </div>
      <div className="w-full relative flex items-center justify-center">
        <QRCode value={invoice} className="w-full" />
        <div className="bg-primary-foreground absolute">
          <LightningIcon className="w-12 h-12" />
        </div>
      </div>
      <div>
        <p className="text-lg font-semibold">
          {new Intl.NumberFormat().format(amount)} sats
        </p>
        <p className="flex flex-col items-center justify-center text-muted-foreground">
          {new Intl.NumberFormat("en-US", {
            currency: "USD",
            style: "currency",
          }).format(fiatAmount)}
        </p>
      </div>
      <div className="flex gap-4 w-full">
        <Button
          onClick={copy}
          variant="outline"
          size={"sm"}
          className="flex-1 flex gap-2 items-center justify-center"
        >
          <CopyIcon className="w-4 h-4 mr-2" />
          Copy
        </Button>
        <>
          {canPayInternally ? (
            <LoadingButton
              loading={isPaying}
              className="whitespace-nowrap"
              size={"sm"}
              onClick={handlePayment}
            >
              Pay and open channel
            </LoadingButton>
          ) : (
            <ExternalLinkButton
              to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/wallet/open-your-first-channel"
              variant="secondary"
              className="flex-1 flex gap-2 items-center justify-center"
            >
              <LightbulbIcon className="w-4 h-4" />
              How to pay
            </ExternalLinkButton>
          )}
        </>
      </div>
    </div>
  );
}
