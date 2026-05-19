import { CreditCardIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { MempoolAlert } from "src/components/MempoolAlert";
import { PayLightningInvoice } from "src/components/PayLightningInvoice";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { LSPTermsDialog } from "src/components/channels/LSPTermsDialog";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { LinkButton } from "src/components/ui/custom/link-button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import {
  Field,
  FieldContent,
  FieldLabel,
  FieldTitle,
} from "src/components/ui/field";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { Separator } from "src/components/ui/separator";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useLSPChannelOffer } from "src/hooks/useLSPChannelOffer";
import { cn } from "src/lib/utils";
import {
  AutoChannelRequest,
  AutoChannelResponse,
  LSPChannelOffer,
  LSPChannelOfferPaymentMethod,
} from "src/types";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request";

import LightningNetworkDarkSVG from "public/images/illustrations/lightning-network-dark.svg";
import LightningNetworkLightSVG from "public/images/illustrations/lightning-network-light.svg";

type FirstChannelProps = {
  variant?: "app" | "setup";
};

const DEFAULT_PAYMENT_METHOD = "wallet" satisfies LSPChannelOfferPaymentMethod;
const PAYMENT_DETAILS_URL = "https://getalby.com/payment_details";

export function FirstChannel({ variant = "app" }: FirstChannelProps) {
  const isSetup = variant === "setup";
  const { data: info, hasChannelManagement } = useInfo();
  const shouldLoadChannelData =
    !info || (!isSetup && !info.albyAccountConnected)
      ? true
      : Boolean(hasChannelManagement && info.albyAccountConnected);
  const { data: channels } = useChannels(true, shouldLoadChannelData);
  const { data: lspChannelOffer } = useLSPChannelOffer(
    shouldLoadChannelData && Boolean(info?.albyAccountConnected)
  );
  const [isLoading, setLoading] = React.useState(false);
  const [showAdvanced, setShowAdvanced] = React.useState(false);
  const [isPublic, setPublic] = React.useState(false);

  const navigate = useNavigate();
  const [invoice, setInvoice] = React.useState<string>();
  const [channelSizeSat, setChannelSizeSat] = React.useState<number>();
  const [currentPaymentMethod, setCurrentPaymentMethod] =
    React.useState<LSPChannelOfferPaymentMethod>(DEFAULT_PAYMENT_METHOD);

  const openingPath = isSetup
    ? "/setup/first-channel/opening"
    : "/channels/first/opening";

  React.useEffect(() => {
    if (channels?.length) {
      navigate(openingPath);
    }
  }, [channels, navigate, openingPath]);

  React.useEffect(() => {
    if (!isSetup && info && !info.albyAccountConnected) {
      navigate("/channels/incoming");
    }
  }, [info, isSetup, navigate]);

  React.useEffect(() => {
    if (isSetup && info && !info.albyAccountConnected) {
      navigate("/home", { replace: true });
    }
  }, [info, isSetup, navigate]);

  React.useEffect(() => {
    if (!lspChannelOffer) {
      return;
    }
    if (
      lspChannelOffer.currentPaymentMethod === "card" ||
      lspChannelOffer.currentPaymentMethod === "wallet"
    ) {
      setCurrentPaymentMethod(lspChannelOffer.currentPaymentMethod);
    }
  }, [lspChannelOffer]);

  if (!info || (shouldLoadChannelData && !channels)) {
    return <Loading />;
  }

  if (isSetup && !hasChannelManagement) {
    return (
      <div className="flex w-full max-w-md flex-col gap-6">
        <TwoColumnLayoutHeader
          title="Your Hub is ready"
          pageTitle="Your Hub is ready"
          description="This wallet setup does not need Lightning channels."
        />
        <LinkButton to="/home" className="w-full justify-center">
          Go to dashboard
        </LinkButton>
      </div>
    );
  }

  if (!info.albyAccountConnected) {
    if (isSetup) {
      return null;
    }
    return <Loading />;
  }

  if (!lspChannelOffer) {
    return <Loading />;
  }

  const showPaymentMethod =
    lspChannelOffer.currentPaymentMethod !== "prepaid" &&
    lspChannelOffer.currentPaymentMethod !== "included";
  const needsPaymentCard =
    currentPaymentMethod === "card" &&
    lspChannelOffer.currentPaymentMethod !== "card";

  async function openChannel() {
    if (!info || !channels || !lspChannelOffer) {
      return;
    }
    if (
      (lspChannelOffer.currentPaymentMethod === "card" ||
        lspChannelOffer.currentPaymentMethod === "wallet") &&
      currentPaymentMethod !== lspChannelOffer.currentPaymentMethod
    ) {
      if (currentPaymentMethod === "card") {
        openLink(PAYMENT_DETAILS_URL);
      }
      toast.error("Payment method incorrectly configured", {
        description: currentPaymentMethod
          ? "Please switch the payment method and confirm the change in your Alby Account settings"
          : "Please choose a payment method",
      });
      return;
    }
    setLoading(true);
    try {
      const newInstantChannelInvoiceRequest: AutoChannelRequest = {
        isPublic,
      };
      const autoChannelResponse = await request<AutoChannelResponse>(
        "/api/alby/auto-channel",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(newInstantChannelInvoiceRequest),
        }
      );
      if (!autoChannelResponse) {
        throw new Error("unexpected auto channel response");
      }

      setInvoice(autoChannelResponse.invoice);
      setChannelSizeSat(autoChannelResponse.channelSizeSat);
    } catch (error) {
      setLoading(false);
      console.error(error);
      toast.error("Something went wrong. Please try again");
    }
  }

  return (
    <>
      {isSetup ? (
        <title>Open Lightning Channel · Alby Hub</title>
      ) : (
        <AppHeader
          pageTitle="Open Lightning Channel"
          title="Open Lightning Channel"
          description="Instantly open a Lightning channel before you start using your Hub."
        />
      )}
      {invoice && channelSizeSat ? (
        <FirstChannelPayment
          channelSizeSat={channelSizeSat}
          invoice={invoice}
          lspChannelOffer={lspChannelOffer}
        />
      ) : (
        <div className="flex flex-col gap-6 max-w-md text-muted-foreground">
          <FirstChannelIntro lspChannelOffer={lspChannelOffer} />

          {showPaymentMethod && (
            <PaymentMethodSelector
              currentPaymentMethod={currentPaymentMethod}
              onPaymentMethodChange={setCurrentPaymentMethod}
            />
          )}

          {needsPaymentCard && <AddPaymentCard />}

          {!showAdvanced && (
            <div className="flex justify-center">
              <Button
                type="button"
                variant="ghost"
                onClick={() => setShowAdvanced(true)}
              >
                Advanced Options
              </Button>
            </div>
          )}

          {showAdvanced && (
            <PublicChannelOption isPublic={isPublic} setPublic={setPublic} />
          )}

          <Separator />

          <ChannelSummary lspChannelOffer={lspChannelOffer} />

          <div className="flex flex-col gap-3">
            <LoadingButton loading={isLoading} onClick={openChannel}>
              {getOpenChannelCta(lspChannelOffer, currentPaymentMethod)}
            </LoadingButton>
            {isSetup && (
              <LinkButton
                to="/home"
                variant="outline"
                className="w-full justify-center"
              >
                Skip for Now
              </LinkButton>
            )}
          </div>

          <p className="text-center text-xs text-muted-foreground">
            By committing, you agree to the{" "}
            <LSPTermsDialog
              contactUrl={lspChannelOffer.lspContactUrl}
              description={lspChannelOffer.lspDescription}
              name={lspChannelOffer.lspName}
              terms={lspChannelOffer.terms}
              trigger={<span className="underline">LSP terms</span>}
            />
            .
          </p>
        </div>
      )}
    </>
  );
}

function FirstChannelIntro({
  lspChannelOffer,
}: {
  lspChannelOffer: LSPChannelOffer;
}) {
  return (
    <>
      <MempoolAlert />
      <div className="flex flex-col gap-2 items-center justify-center py-2 text-center">
        <h1 className="font-semibold text-2xl text-foreground">
          Open Lightning Channel
        </h1>
        <p className="text-sm text-muted-foreground">
          Instantly open a lightning channel to another network node with our
          partner{" "}
          <ExternalLink
            to={lspChannelOffer.lspContactUrl}
            className="text-foreground underline"
          >
            {lspChannelOffer.lspName}
          </ExternalLink>
          . Once opened, you'll be able to send and receive bitcoin instantly.{" "}
          <LearnMoreDialog />
        </p>
      </div>
    </>
  );
}

function PaymentMethodSelector({
  currentPaymentMethod,
  onPaymentMethodChange,
}: {
  currentPaymentMethod: LSPChannelOfferPaymentMethod;
  onPaymentMethodChange(paymentMethod: LSPChannelOfferPaymentMethod): void;
}) {
  return (
    <div className="flex flex-col gap-3">
      <h2 className="font-medium text-sm text-foreground">Payment method</h2>
      <RadioGroup
        value={currentPaymentMethod}
        onValueChange={(newPaymentMethod: LSPChannelOfferPaymentMethod) =>
          onPaymentMethodChange(newPaymentMethod)
        }
      >
        <FieldLabel htmlFor="wallet" className="text-foreground">
          <Field orientation="horizontal">
            <RadioGroupItem
              value={"wallet" satisfies LSPChannelOfferPaymentMethod}
              id="wallet"
            />
            <FieldContent>
              <FieldTitle className="text-foreground">Bitcoin</FieldTitle>
            </FieldContent>
          </Field>
        </FieldLabel>
        <FieldLabel htmlFor="card" className="text-foreground">
          <Field orientation="horizontal">
            <RadioGroupItem
              value={"card" satisfies LSPChannelOfferPaymentMethod}
              id="card"
            />
            <FieldContent>
              <FieldTitle className="text-foreground">
                Credit / Debit Card
              </FieldTitle>
            </FieldContent>
          </Field>
        </FieldLabel>
      </RadioGroup>
    </div>
  );
}

function AddPaymentCard() {
  return (
    <div className="border rounded-md p-4 flex items-start gap-4">
      <div className="size-8 rounded-sm border bg-muted flex items-center justify-center shrink-0">
        <CreditCardIcon className="size-4" />
      </div>
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <p className="text-sm font-medium text-foreground">
          Add a payment card
        </p>
        <p className="text-sm text-muted-foreground">
          Save your card in Alby Account settings to pay for the channel.
        </p>
      </div>
      <Button
        type="button"
        variant="secondary"
        onClick={() => openLink(PAYMENT_DETAILS_URL)}
      >
        Add Card
      </Button>
    </div>
  );
}

function PublicChannelOption({
  isPublic,
  setPublic,
}: {
  isPublic: boolean;
  setPublic(isPublic: boolean): void;
}) {
  return (
    <div className="flex items-start gap-2">
      <Checkbox
        id="public-channel"
        checked={isPublic}
        onCheckedChange={(checked) => setPublic(checked === true)}
      />
      <div className="grid gap-1.5 leading-none">
        <Label
          htmlFor="public-channel"
          className="cursor-pointer text-foreground"
        >
          Public channel
        </Label>
        <p className="text-sm text-muted-foreground">
          Not recommended for most users.{" "}
          <ExternalLink
            className="underline"
            to="https://guides.getalby.com/user-guide/alby-hub/faq/should-i-open-a-private-or-public-channel"
          >
            Learn more
          </ExternalLink>
        </p>
      </div>
    </div>
  );
}

function ChannelSummary({
  lspChannelOffer,
}: {
  lspChannelOffer: LSPChannelOffer;
}) {
  return (
    <div className="grid gap-3">
      <div className="flex items-start justify-between gap-4 text-sm">
        <p className="text-muted-foreground">
          You'll be able to receive up to:
        </p>
        <FormattedFiatAmount
          amountSat={lspChannelOffer.lspBalanceSat}
          className="font-medium text-foreground"
          showApprox
        />
      </div>
      <div className="flex items-start justify-between gap-4 text-sm">
        <p className="text-muted-foreground">Total channel cost:</p>
        <p className="font-medium text-foreground">
          <span
            className={cn(
              lspChannelOffer.currentPaymentMethod === "included" &&
                "line-through"
            )}
          >
            {new Intl.NumberFormat(undefined, {
              style: "currency",
              currency: "USD",
            }).format(lspChannelOffer.feeTotalUsd / 100)}
          </span>
          {lspChannelOffer.currentPaymentMethod === "included" && (
            <span> $0.00</span>
          )}
        </p>
      </div>
    </div>
  );
}

function FirstChannelPayment({
  channelSizeSat,
  invoice,
  lspChannelOffer,
}: {
  channelSizeSat: number;
  invoice: string;
  lspChannelOffer: LSPChannelOffer;
}) {
  return (
    <div className="flex w-full max-w-md flex-col items-center justify-center gap-6">
      <h1 className="text-center text-2xl font-semibold">Open First Channel</h1>
      <PayLightningInvoice
        invoice={invoice}
        variant="first-channel"
        incomingLiquiditySat={channelSizeSat}
        lspTerms={{
          name: lspChannelOffer.lspName,
          description: lspChannelOffer.lspDescription,
          contactUrl: lspChannelOffer.lspContactUrl,
          terms: lspChannelOffer.terms,
        }}
      />
    </div>
  );
}

function LearnMoreDialog() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <button
          type="button"
          className="text-foreground underline underline-offset-2"
        >
          Learn More
        </button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Lightning Network Channels</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          <p className="text-sm text-muted-foreground">
            Lightning channels allow you to send and receive bitcoin instantly
            on the lightning network. Each channel has a certain capacity of how
            much bitcoin it can receive.
          </p>
          <p className="text-sm text-muted-foreground">
            Alby Hub opens lightning channels to selected liquidity providers
            (LSPs) for optimal network connectivity and reliability.
          </p>
          <div className="overflow-hidden rounded-lg p-4">
            <img
              src={LightningNetworkDarkSVG}
              className="hidden w-full dark:block"
              alt=""
            />
            <img
              src={LightningNetworkLightSVG}
              className="w-full dark:hidden"
              alt=""
            />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function getOpenChannelCta(
  lspChannelOffer: LSPChannelOffer,
  currentPaymentMethod: LSPChannelOfferPaymentMethod
) {
  if (lspChannelOffer.currentPaymentMethod === "prepaid") {
    return "Review Order";
  }
  if (lspChannelOffer.currentPaymentMethod === "included") {
    return "Open Channel";
  }
  if (currentPaymentMethod === "card") {
    return `Pay ${new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: "USD",
    }).format(lspChannelOffer.feeTotalUsd / 100)} and Open Channel`;
  }
  return "Review Order";
}
