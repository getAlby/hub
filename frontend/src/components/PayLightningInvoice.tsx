import { Invoice, fiat } from "@getalby/lightning-tools";
import { CopyIcon, LightbulbIcon } from "lucide-react";
import React from "react";
import { LightningIcon } from "src/components/icons/Lightning";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";

type PayLightningInvoiceProps = {
  invoice: string;
};

export function PayLightningInvoice({ invoice }: PayLightningInvoiceProps) {
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

  return (
    <div className="w-96 flex flex-col gap-6 p-6 items-center justify-center">
      <div className="flex items-center justify-center gap-2 text-muted-foreground">
        <Loading variant="loader" />
        <p>Waiting for lightning payment...</p>
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
        <p className="flex flex-col items-center justify-center">
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
          className="flex-1 flex gap-2 items-center justify-center"
        >
          <CopyIcon className="w-4 h-4 mr-2" />
          Copy Invoice
        </Button>
        <ExternalLinkButton
          to="https://guides.getalby.com/user-guide/alby-hub/wallet/open-your-first-channel"
          variant="secondary"
          className="flex-1 flex gap-2 items-center justify-center"
        >
          <LightbulbIcon className="w-4 h-4" /> How to pay
        </ExternalLinkButton>
      </div>
    </div>
  );
}
