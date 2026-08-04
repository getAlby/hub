import { Invoice, getFiatValue } from "@getalby/lightning-tools";
import { CopyIcon, ExternalLinkIcon } from "lucide-react";
import React from "react";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import { copyToClipboard } from "src/lib/clipboard";

type PayLightningInvoiceProps = {
  invoice: string;
};

export function PayLightningInvoice({ invoice }: PayLightningInvoiceProps) {
  const amountSat = new Invoice({
    pr: invoice,
  }).satoshi;
  const [fiatAmount, setFiatAmount] = React.useState(0);
  React.useEffect(() => {
    getFiatValue({ satoshi: amountSat, currency: "USD" }).then((fiatAmount) =>
      setFiatAmount(fiatAmount)
    );
  }, [amountSat]);
  const copy = () => {
    copyToClipboard(invoice);
  };

  return (
    <div className="flex w-full flex-col items-center justify-center gap-6">
      <div className="flex items-center justify-center gap-2 font-semibold leading-none">
        <Loading variant="loader" />
        <p>Waiting for Payment...</p>
      </div>
      <div className="relative flex w-full items-center justify-center">
        <QRCode value={invoice} className="w-full" paymentType="lightning" />
      </div>
      <div>
        <p className="text-lg font-semibold">
          <FormattedBitcoinAmount amountMsat={amountSat * 1000} />
        </p>
        <p className="flex flex-col items-center justify-center">
          {new Intl.NumberFormat("en-US", {
            currency: "USD",
            style: "currency",
          }).format(fiatAmount)}
        </p>
      </div>
      <div className="flex w-full flex-col gap-3">
        <Button onClick={copy} variant="secondary" className="w-full">
          <CopyIcon />
          Copy Invoice
        </Button>
        <FixedFloatButton
          to="BTCLN"
          address={invoice}
          className="w-full"
          variant="outline"
        >
          <ExternalLinkIcon className="size-4" />
          Pay with Crypto
        </FixedFloatButton>
      </div>
    </div>
  );
}
