import {
  ArrowLeftIcon,
  CopyIcon,
  LinkIcon,
  PlusIcon,
  ReceiptTextIcon,
} from "lucide-react";
import TickSVG from "public/images/illustrations/tick.svg";
import React from "react";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { CurrencyInputField } from "src/components/CurrencyInputField";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import QRCode from "src/components/QRCode";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";
import { useTransaction } from "src/hooks/useTransaction";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import { CreateInvoiceRequest, Transaction } from "src/types";
import { request } from "src/utils/request";

export default function ReceiveInvoice() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: me } = useAlbyMe();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();

  const [isLoading, setLoading] = React.useState(false);
  const [amountSat, setAmountSat] = React.useState<string>("");
  const [description, setDescription] = React.useState<string>("");
  const [transaction, setTransaction] = React.useState<Transaction | null>(
    null
  );
  const { data: invoiceData } = useTransaction(
    transaction ? transaction.paymentHash : "",
    true
  );
  const paymentDone = !!invoiceData?.settledAt;
  const jitChannelsEnabled = !!info?.jitChannelsEnabled;
  const configuredLsps2Source = info?.jitChannelsLiquiditySource;
  const lsps2Source = jitChannelsEnabled ? configuredLsps2Source : undefined;
  const lsps2MinimumPaymentSizeSat = React.useMemo(() => {
    if (jitChannelsEnabled && info?.jitChannelsMinPaymentSizeMsat) {
      return Math.ceil(info.jitChannelsMinPaymentSizeMsat / 1000);
    }
    return undefined;
  }, [info?.jitChannelsMinPaymentSizeMsat, jitChannelsEnabled]);
  // only enforce the minimum on the input when the user has no channels yet -
  // their first channel must meet the minimum size.
  const jitMinimumReceiveSat = channels?.length
    ? undefined
    : lsps2MinimumPaymentSizeSat;
  const lsps2MaximumPaymentSizeSat = React.useMemo(() => {
    if (jitChannelsEnabled && info?.jitChannelsMaxPaymentSizeMsat) {
      return Math.floor(info.jitChannelsMaxPaymentSizeMsat / 1000);
    }
    return undefined;
  }, [info?.jitChannelsMaxPaymentSizeMsat, jitChannelsEnabled]);
  const jitMaximumReceiveSat =
    hasChannelManagement && lsps2Source
      ? lsps2MaximumPaymentSizeSat
      : !lsps2Source && hasChannelManagement
        ? balances?.lightning.totalReceivableSat
        : undefined;
  const totalReceivableMsat = balances?.lightning.totalReceivableMsat ?? 0;
  const requestedAmountMsat = +amountSat * 1000 || transaction?.amountMsat || 0;
  const isNearReceivingCapacity =
    !!hasChannelManagement && requestedAmountMsat >= 0.8 * totalReceivableMsat;
  const isJitReceiveInvoice =
    !!hasChannelManagement &&
    !!lsps2Source &&
    !!transaction &&
    transaction.amountMsat > totalReceivableMsat;
  const displayedJitFeeMsat = paymentDone
    ? (invoiceData?.feesPaidMsat ?? 0)
    : (transaction?.feesPaidMsat ?? 0);

  if (!balances || !info || (info.albyAccountConnected && !me)) {
    return <Loading />;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    try {
      setLoading(true);
      const invoice = await request<Transaction>("/api/invoices", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          amountMsat: (parseInt(amountSat) || 0) * 1000,
          description,
        } as CreateInvoiceRequest),
      });

      if (invoice) {
        setTransaction(invoice);
        setAmountSat("");
        setDescription("");
        toast("Successfully created invoice");
      }
    } catch (e) {
      const requestedAmountSat = parseInt(amountSat) || 0;
      // the user already has channels but this amount exceeds their receiving
      // capacity (so a new channel is needed) and is below the LSP's minimum
      // channel size - the receive may have failed because the amount was too
      // small to open a second channel, so add a hint alongside the error.
      const likelyTooSmallForNewChannel =
        jitChannelsEnabled &&
        !!channels?.length &&
        !!lsps2MinimumPaymentSizeSat &&
        requestedAmountSat < lsps2MinimumPaymentSizeSat &&
        requestedAmountSat * 1000 > totalReceivableMsat;
      let description = "" + e;
      if (likelyTooSmallForNewChannel) {
        description += `\n\nThis amount is over your receiving capacity and may be too small to open a new Lightning channel. Try receiving at least ${new Intl.NumberFormat().format(
          lsps2MinimumPaymentSizeSat as number
        )} sats, or lower the amount to fit your current capacity.`;
      }
      toast.error("Failed to create invoice", {
        description,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const copy = () => {
    copyToClipboard(transaction?.invoice as string);
  };

  const newChannelFeeAlert = (
    <p className="text-sm text-muted-foreground text-center">
      Includes a <FormattedBitcoinAmount amountMsat={displayedJitFeeMsat} />{" "}
      channel fee.{" "}
      <ExternalLink
        to="https://guides.getalby.com/user-guide/alby-hub/faq/what-are-just-in-time-channels"
        className="underline"
      >
        Learn more
      </ExternalLink>
    </p>
  );

  return (
    <div className="grid gap-5">
      <AppHeader
        pageTitle={transaction ? "Lightning Invoice" : "Create Invoice"}
        title={transaction ? "Lightning Invoice" : "Create Invoice"}
      />
      <div className="flex flex-col md:flex-row gap-12">
        <div className="w-full md:max-w-lg grid gap-6">
          {!lsps2Source && !transaction && isNearReceivingCapacity && (
            <LowReceivingCapacityAlert />
          )}
          <div>
            {transaction ? (
              <Card>
                {!paymentDone ? (
                  <>
                    <CardHeader>
                      <CardTitle className="flex justify-center">
                        <Loading className="size-4 mr-2" />
                        <p>Waiting for payment</p>
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col items-center gap-6">
                      <QRCode value={transaction.invoice} />
                      <div className="flex flex-col gap-1 items-center">
                        <p className="text-2xl font-medium slashed-zero">
                          <FormattedBitcoinAmount
                            amountMsat={transaction.amountMsat}
                          />
                        </p>
                        <FormattedFiatAmount
                          amountSat={transaction.amountSat}
                          className="text-xl"
                        />
                      </div>
                      {isJitReceiveInvoice && displayedJitFeeMsat >= 1000 && (
                        <div className="w-full">{newChannelFeeAlert}</div>
                      )}
                    </CardContent>
                    <CardFooter className="flex flex-col gap-2">
                      <Button
                        className="w-full"
                        onClick={copy}
                        variant="outline"
                      >
                        <CopyIcon className="w-4 h-4 mr-2" />
                        Copy Invoice
                      </Button>
                    </CardFooter>
                  </>
                ) : (
                  <>
                    <CardHeader>
                      <CardTitle className="text-center">
                        Payment Received
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col items-center gap-6">
                      <img src={TickSVG} className="w-48" />
                      <div className="flex flex-col gap-1 items-center">
                        <p className="text-2xl font-medium slashed-zero">
                          <FormattedBitcoinAmount
                            amountMsat={transaction.amountMsat}
                          />
                        </p>
                        <FormattedFiatAmount
                          amountSat={transaction.amountSat}
                          className="text-xl"
                        />
                      </div>
                    </CardContent>
                    <CardFooter className="flex flex-col gap-2 pt-2">
                      <Button
                        onClick={() => {
                          setTransaction(null);
                        }}
                        variant="outline"
                        className="w-full"
                      >
                        <PlusIcon className="w-4 h-4 mr-2" />
                        Create Another Invoice
                      </Button>
                      <LinkButton
                        to="/wallet"
                        variant="link"
                        className="w-full"
                      >
                        <ArrowLeftIcon className="w-4 h-4 mr-2" />
                        Back to Wallet
                      </LinkButton>
                    </CardFooter>
                  </>
                )}
              </Card>
            ) : (
              <form onSubmit={handleSubmit} className="grid gap-6">
                <CurrencyInputField
                  id="amount"
                  valueSat={amountSat}
                  onValueSatChange={setAmountSat}
                  minSat={jitMinimumReceiveSat ?? 1}
                  onInvalid={(e) => {
                    if (
                      jitMinimumReceiveSat &&
                      e.currentTarget.validity.rangeUnderflow
                    ) {
                      e.currentTarget.setCustomValidity(
                        `You need to receive at least ${new Intl.NumberFormat().format(
                          jitMinimumReceiveSat
                        )} sats to open your first lightning channel`
                      );
                    } else if (
                      jitMaximumReceiveSat &&
                      e.currentTarget.validity.rangeOverflow
                    ) {
                      e.currentTarget.setCustomValidity(
                        lsps2Source
                          ? `This JIT channel setup supports receiving at most ${new Intl.NumberFormat().format(
                              jitMaximumReceiveSat
                            )} sats in a single payment`
                          : `You can receive at most ${new Intl.NumberFormat().format(
                              jitMaximumReceiveSat
                            )} sats with your current capacity`
                      );
                    } else {
                      e.currentTarget.setCustomValidity("");
                    }
                  }}
                  maxSat={jitMaximumReceiveSat}
                  autoFocus
                  contextRows={
                    hasChannelManagement && !lsps2Source && jitMaximumReceiveSat
                      ? [
                          {
                            label: "Receive limit",
                            amountSat: jitMaximumReceiveSat,
                          },
                        ]
                      : undefined
                  }
                />
                <div className="grid gap-2">
                  <Label htmlFor="description">Description</Label>
                  <Input
                    id="description"
                    type="text"
                    value={description}
                    placeholder="For e.g. who is sending this payment?"
                    onChange={(e) => {
                      setDescription(e.target.value);
                    }}
                  />
                </div>
                <LoadingButton
                  className={cn(
                    "w-full",
                    info?.albyAccountConnected &&
                      me?.lightning_address &&
                      "md:w-fit"
                  )}
                  loading={isLoading}
                  type="submit"
                  disabled={!amountSat}
                >
                  Create Invoice
                </LoadingButton>
                {(!info?.albyAccountConnected || !me?.lightning_address) && (
                  <Accordion type="single" collapsible className="w-full">
                    <AccordionItem value="more-options">
                      <AccordionTrigger>
                        View other ways to receive
                      </AccordionTrigger>
                      <AccordionContent className="flex flex-col gap-2">
                        {!info?.albyAccountConnected && info.supportsBolt12 && (
                          <LinkButton
                            to="/wallet/receive/offer"
                            variant="outline"
                            className="w-full"
                          >
                            <ReceiptTextIcon className="h-4 w-4" />
                            Lightning Offer
                          </LinkButton>
                        )}
                        <LinkButton
                          to="/wallet/receive/onchain"
                          variant="outline"
                          className="w-full"
                        >
                          <LinkIcon className="h-4 w-4" />
                          Receive from On-chain / Other Cryptocurrency
                        </LinkButton>
                      </AccordionContent>
                    </AccordionItem>
                  </Accordion>
                )}
              </form>
            )}
          </div>
        </div>
        {!transaction &&
          (!info?.albyAccountConnected || !me?.lightning_address) && (
            <LightningAddressCard />
          )}
      </div>
    </div>
  );
}

function LightningAddressCard() {
  return (
    <Card className="w-full self-start">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="font-semibold text-lg">
          Get Your Free Lightning Address
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-3 text-muted-foreground">
          <p>
            Create free Alby Account and link it with your Alby Hub to get a
            convenient <span className="text-foreground">@getalby.com</span>{" "}
            lightning address and other perks:
          </p>
          <ul className="flex flex-col gap-1">
            <li>• Lightning address & Nostr identifier,</li>
            <li>• Personal tipping page,</li>
            <li>• Access to podcasting 2.0 apps,</li>
            <li>• Buy bitcoin directly to your wallet,</li>
            <li>• Useful email Alby Hub notifications.</li>
          </ul>
        </div>
      </CardContent>
      <CardFooter className="flex justify-end">
        <ExternalLinkButton
          to="https://getalby.com/auth/users/new"
          variant="secondary"
        >
          Get Alby Account
        </ExternalLinkButton>
      </CardFooter>
    </Card>
  );
}
