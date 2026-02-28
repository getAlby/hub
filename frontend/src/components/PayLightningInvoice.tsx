import { Invoice, getFiatValue } from "@getalby/lightning-tools";
import { CopyIcon, ExternalLinkIcon } from "lucide-react";
import React from "react";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { LightningIcon } from "src/components/icons/Lightning";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import { useBitcoinMaxiMode } from "src/hooks/useBitcoinMaxiMode";
import { copyToClipboard } from "src/lib/clipboard";

type PayLightningInvoiceProps = {
  invoice: string;
};

export function PayLightningInvoice({ invoice }: PayLightningInvoiceProps) {
  const { bitcoinMaxiMode } = useBitcoinMaxiMode();
  const amount = new Invoice({
    pr: invoice,
  }).satoshi;
  const [fiatAmount, setFiatAmount] = React.useState(0);
  React.useEffect(() => {
    getFiatValue({ satoshi: amount, currency: "USD" }).then((fiatAmount) =>
      setFiatAmount(fiatAmount)
    );
  }, [amount]);
  const copy = () => {
    copyToClipboard(invoice);
  };

  return (
    <div className="w-96 flex flex-col gap-6 p-6 items-center justify-center">
      <div className="flex items-center justify-center gap-2 text-muted-foreground">
        <Loading variant="loader" />
        <p>Waiting for lightning payment...</p>
      </div>
      <div className="w-full relative flex items-center justify-center">
        <QRCode value={invoice} className="w-full" />
        <div className="bg-white absolute rounded-full p-1">
          <LightningIcon className="w-12 h-12" />
        </div>
      </div>
      <div>
        <p className="text-lg font-semibold">
          <FormattedBitcoinAmount amount={amount * 1000} />
        </p>
        <p className="flex flex-col items-center justify-center">
          {new Intl.NumberFormat("en-US", {
            currency: "USD",
            style: "currency",
          }).format(fiatAmount)}
        </p>
      </div>
      <div className="flex flex-col gap-2 w-full">
        <Button
          onClick={copy}
          variant="outline"
          className="flex-1 flex gap-2 items-center justify-center"
        >
          <CopyIcon />
          Copy Invoice
        </Button>
        {!bitcoinMaxiMode && (
          <FixedFloatButton
            to="BTCLN"
            address={invoice}
            className="flex-1 flex gap-2 items-center justify-center"
            variant="secondary"
          >
            Pay with other Cryptocurrency
            <ExternalLinkIcon className="size-4" />
          </FixedFloatButton>
        )}
      </div>
    </div>
  );
}
