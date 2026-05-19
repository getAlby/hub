import { Invoice, getFiatValue } from "@getalby/lightning-tools";
import {
  ArrowRightLeftIcon,
  CopyIcon,
  ExternalLinkIcon,
  InfoIcon,
} from "lucide-react";
import React from "react";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { LightningIcon } from "src/components/icons/Lightning";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import { LSPTermsDialog } from "src/components/channels/LSPTermsDialog";
import { Button } from "src/components/ui/button";
import { LinkButton } from "src/components/ui/custom/link-button";
import { Separator } from "src/components/ui/separator";
import { copyToClipboard } from "src/lib/clipboard";

type PayLightningInvoiceProps = {
  invoice: string;
  variant?: "default" | "first-channel";
  amountSat?: number;
  incomingLiquiditySat?: number;
  lspTerms?: {
    name: string;
    description: string;
    contactUrl: string;
    terms?: string;
  };
};

export function PayLightningInvoice({
  invoice,
  variant = "default",
  amountSat: amountSatProp,
  incomingLiquiditySat,
  lspTerms,
}: PayLightningInvoiceProps) {
  const amountSat =
    amountSatProp ??
    new Invoice({
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

  if (variant === "first-channel") {
    return (
      <div className="flex w-full max-w-md flex-col items-center gap-6 rounded-xl border bg-background p-6">
        <div className="flex items-center justify-center gap-2 font-semibold">
          <Loading variant="loader" />
          <p>Waiting for Lightning Payment</p>
        </div>
        <div className="relative flex w-64 items-center justify-center">
          <QRCode value={invoice} className="w-full" />
          <div className="absolute rounded-full bg-white p-2">
            <LightningIcon className="h-14 w-14" />
          </div>
        </div>
        <div className="text-center">
          <p className="text-lg font-semibold">
            <FormattedBitcoinAmount amountMsat={amountSat * 1000} />
          </p>
          <p className="text-muted-foreground">
            {new Intl.NumberFormat("en-US", {
              currency: "USD",
              style: "currency",
            }).format(fiatAmount)}
          </p>
        </div>

        {lspTerms && (
          <p className="text-center text-sm text-muted-foreground">
            By purchasing, you agree to{" "}
            <LSPTermsDialog
              contactUrl={lspTerms.contactUrl}
              description={lspTerms.description}
              name={lspTerms.name}
              terms={lspTerms.terms}
              trigger={
                <span className="underline underline-offset-2">
                  {lspTerms.name} Terms & Conditions
                </span>
              }
            />
            .
          </p>
        )}

        <div className="flex w-full flex-col gap-3">
          <Button
            onClick={copy}
            variant="secondary"
            className="flex items-center justify-center gap-2"
          >
            <CopyIcon />
            Copy Invoice
          </Button>
          <LinkButton
            to="/channels/outgoing"
            variant="outline"
            className="w-full justify-center gap-2"
          >
            Open Channel with On-Chain
            <ArrowRightLeftIcon className="size-4" />
          </LinkButton>
          <FixedFloatButton
            to="BTCLN"
            address={invoice}
            className="flex items-center justify-center gap-2"
            variant="outline"
          >
            Pay with Other Crypto
            <ExternalLinkIcon className="size-4" />
          </FixedFloatButton>
        </div>

        <Separator />

        <div className="grid w-full gap-3 text-sm">
          {incomingLiquiditySat !== undefined && (
            <div className="flex items-center justify-between gap-4">
              <p className="text-muted-foreground">Incoming liquidity</p>
              <FormattedBitcoinAmount
                amountMsat={incomingLiquiditySat * 1000}
              />
            </div>
          )}
          <div className="flex items-center justify-between gap-4">
            <p className="text-muted-foreground">Channel duration</p>
            <div className="flex items-center gap-2">
              <InfoIcon className="size-3 text-muted-foreground" />
              <span>at least 3 months</span>
            </div>
          </div>
        </div>
      </div>
    );
  }

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
          <FormattedBitcoinAmount amountMsat={amountSat * 1000} />
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
        <FixedFloatButton
          to="BTCLN"
          address={invoice}
          className="flex-1 flex gap-2 items-center justify-center"
          variant="secondary"
        >
          Pay with other Cryptocurrency
          <ExternalLinkIcon className="size-4" />
        </FixedFloatButton>
      </div>
    </div>
  );
}
