import {
  CopyIcon,
  ExternalLinkIcon,
  LinkIcon,
  QrCodeIcon,
  ReceiptTextIcon,
  RefreshCwIcon,
  ZapIcon,
} from "lucide-react";
import { FixedFloatButton } from "src/components/FixedFloatButton";
import Loading from "src/components/Loading";
import LottieLoading from "src/components/LottieLoading";
import LottieSuccess from "src/components/LottieSuccess";
import OnchainAddressDisplay from "src/components/OnchainAddressDisplay";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { Separator } from "src/components/ui/separator";

const DUMMY_INVOICE =
  "lnbc100u1p5hkvrndpz2pshjmt9de6zqmmxyqcnqvpsxqs8xct5wvnp4qwmtpr4p72ms7gnq3pkfk2876y2mscpp5hpd6h7t023cf3q8d06y9slqcnkydffgzun9th5vjm62nsw8wssgqsp5dfddw9ezn93u7g9xmzh4q74kmwxlf0gxgx8c5e4cuu2ce3eapmgq9qyysgqcqzp2xqyz5vqp0t02p3882uhqsz0qf56jgy6mrf2523tudqnf5d2f6f83ud3krd9tu4zkd4yzwyc74acprnvz2853yf9lc89n90sy3r0lvckyy59racq40428t";
const DUMMY_OFFER =
  "lno1qgsqv3jhxapqypuxzcm99ps8yctjyp4x2um5wf5x5cnxxmmd9akxuatjd3cz7mrww4exc6t0dcsz6un9w3skuatnv5s8gmmnd96xser9v9ehg0";
const DUMMY_LIGHTNING_ADDRESS = "review@getalby.com";
const DUMMY_ONCHAIN_ADDRESS = "bc1qrevieww56s59x9h4n2s2l2n6r4e3m2q5p9y7tx0p";
const DUMMY_TX_ID =
  "9f0c6b0a4f6bce34b1d0a6d1a31ad71f4e9d34e7a2c3b7fd94ab67f2a4c0dd5e";
const DUMMY_POS_URL =
  "https://pos.albylabs.com/#/?nwc=review-only-dummy-pairing&label=Internal%20Review%20PoS";

export function PaymentComponents() {
  return (
    <div className="grid min-w-0 gap-5">
      <header className="grid gap-1">
        <h1 className="text-2xl font-semibold">Payment Component Review</h1>
        <p className="max-w-3xl text-sm text-muted-foreground">
          Dummy QR payment states for internal visual checks. These examples do
          not create invoices, swaps, addresses, or payments.
        </p>
      </header>

      <div className="grid min-w-0 gap-6 xl:grid-cols-3">
        <ReviewSection
          title="Lightning receive"
          description="Wallet receive QR states."
        >
          <LightningAddressCard />
          <LightningInvoiceWaitingCard />
          <LightningInvoicePaidCard />
          <LightningOfferCard />
        </ReviewSection>

        <ReviewSection
          title="On-chain receive"
          description="Address, pending confirmation, and received states."
        >
          <OnchainAddressCard />
          <OnchainPendingCard />
          <OnchainSuccessCard />
        </ReviewSection>

        <ReviewSection
          title="Flow integrations"
          description="Channel funding, swap, and PoS QR surfaces."
        >
          <ComponentFrame title="Channel order lightning invoice">
            <ChannelOrderLightningInvoiceCard />
          </ComponentFrame>
          <ComponentFrame title="FixedFloat invoice waiting">
            <FixedFloatWaitingCard />
          </ComponentFrame>
          <ComponentFrame title="FixedFloat invoice received">
            <FixedFloatSuccessCard />
          </ComponentFrame>
          <BuzzPayCard />
        </ReviewSection>
      </div>
    </div>
  );
}

function ReviewSection({
  title,
  description,
  children,
}: {
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <section className="grid min-w-0 content-start gap-4">
      <div>
        <h2 className="text-xl font-semibold">{title}</h2>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      {children}
    </section>
  );
}

function ComponentFrame({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <Card className="min-w-0">
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex justify-center">{children}</CardContent>
    </Card>
  );
}

function ChannelOrderLightningInvoiceCard() {
  return (
    <div className="flex w-full flex-col items-center justify-center gap-6">
      <div className="flex items-center justify-center gap-2 font-semibold leading-none">
        <Loading variant="loader" />
        <p>Waiting for Payment...</p>
      </div>
      <div className="relative flex w-full items-center justify-center">
        <QRCode
          value={DUMMY_INVOICE}
          className="w-full"
          paymentType="lightning"
        />
      </div>
      <PaymentAmount amountSat={10000} />
      <div className="flex w-full flex-col gap-3">
        <Button variant="secondary" className="w-full">
          <CopyIcon />
          Copy Invoice
        </Button>
        <FixedFloatButton
          to="BTCLN"
          address={DUMMY_INVOICE}
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

function FixedFloatWaitingCard() {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-center">Waiting for Payment...</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <LottieLoading size={288} />
        <PaymentAmount amountSat={10000} />
      </CardContent>
      <CardFooter className="flex flex-col gap-3 pt-2">
        <Button type="button" className="w-full" variant="secondary">
          <CopyIcon className="size-4" />
          Copy Invoice
        </Button>
        <FixedFloatButton
          to="BTCLN"
          address={DUMMY_INVOICE}
          variant="outline"
          className="w-full"
        >
          <ExternalLinkIcon className="size-4" />
          Open Fixed Float
        </FixedFloatButton>
      </CardFooter>
    </Card>
  );
}

function FixedFloatSuccessCard() {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-center">Transaction Received!</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <LottieSuccess />
        <PaymentAmount amountSat={10000} />
      </CardContent>
      <CardFooter className="flex flex-col gap-3 pt-2">
        <Button type="button" variant="outline" className="w-full">
          Receive Another Payment
        </Button>
        <Button variant="link" className="w-full">
          Back to Wallet
        </Button>
      </CardFooter>
    </Card>
  );
}

function LightningAddressCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <QrCodeIcon className="size-4" />
          Lightning address
        </CardTitle>
        <CardDescription>QR shown on the receive overview.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <QRCode
          value={DUMMY_LIGHTNING_ADDRESS}
          className="h-auto w-full"
          frameType="lightning"
        />
        <p className="break-all text-center text-lg font-medium">
          {DUMMY_LIGHTNING_ADDRESS}
        </p>
      </CardContent>
      <CardFooter className="flex flex-col gap-3 pt-2">
        <Button variant="secondary" className="w-full">
          <CopyIcon className="size-4" /> Copy Lightning Address
        </Button>
        <Separator className="my-4" />
        <Button variant="outline" className="w-full">
          <ZapIcon className="size-4" />
          Create Invoice
        </Button>
        <Button variant="outline" className="w-full">
          <ReceiptTextIcon className="size-4" />
          Lightning Offer
        </Button>
        <Button variant="outline" className="w-full">
          <LinkIcon className="size-4" />
          Receive on-chain/crypto
        </Button>
      </CardFooter>
    </Card>
  );
}

function LightningInvoiceWaitingCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-center gap-2 text-base">
          <Loading variant="loader" />
          Waiting for Payment...
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <QRCode value={DUMMY_INVOICE} paymentType="lightning" />
        <PaymentAmount amountSat={10000} />
      </CardContent>
      <CardFooter className="flex flex-col gap-3">
        <Button className="w-full" variant="outline">
          <CopyIcon className="size-4" />
          Copy Invoice
        </Button>
      </CardFooter>
    </Card>
  );
}

function LightningInvoicePaidCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-center text-base">
          Payment Received
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <LottieSuccess />
        <PaymentAmount amountSat={10000} />
      </CardContent>
      <CardFooter className="flex flex-col gap-3 pt-2">
        <Button variant="outline" className="w-full">
          <ZapIcon className="size-4" />
          Create Another Invoice
        </Button>
        <Button variant="link" className="w-full">
          Back to Wallet
        </Button>
      </CardFooter>
    </Card>
  );
}

function LightningOfferCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-center text-base">Lightning Offer</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <QRCode
          value={DUMMY_OFFER}
          className="w-full"
          paymentType="lightning"
        />
        <p className="my-2 text-muted-foreground">Review-only offer</p>
      </CardContent>
      <CardFooter className="flex flex-col gap-3">
        <Button className="w-full" variant="secondary">
          <CopyIcon className="size-4" />
          Copy Offer
        </Button>
        <Button className="w-full" variant="outline">
          <RefreshCwIcon className="size-4" />
          New Offer
        </Button>
      </CardFooter>
    </Card>
  );
}

function OnchainAddressCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">On-chain address</CardTitle>
        <CardDescription>
          Deposit QR used by wallet and channel flows.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <QRCode value={DUMMY_ONCHAIN_ADDRESS} paymentType="onchain" />
        <div className="flex max-w-64 flex-wrap items-center justify-center gap-2">
          <OnchainAddressDisplay address={DUMMY_ONCHAIN_ADDRESS} />
        </div>
      </CardContent>
      <CardFooter className="flex flex-col gap-3 pt-2">
        <Button className="w-full" variant="secondary">
          <CopyIcon className="size-4" />
          Copy Address
        </Button>
        <Button className="w-full" variant="outline">
          <RefreshCwIcon className="size-4" />
          New Address
        </Button>
        <Separator className="my-4" />
        <FixedFloatButton
          to="BTC"
          address={DUMMY_ONCHAIN_ADDRESS}
          className="w-full"
          variant="outline"
        >
          <ExternalLinkIcon className="size-4" />
          Top Up with Crypto
        </FixedFloatButton>
      </CardFooter>
    </Card>
  );
}

function OnchainPendingCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-center text-base">
          Waiting for On-chain Confirmation...
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-4">
        <LottieLoading size={288} />
        <PaymentAmount amountSat={25000} />
      </CardContent>
      <CardFooter className="flex flex-col gap-3 pt-2">
        <ExternalLinkButton
          to={`https://mempool.space/tx/${DUMMY_TX_ID}`}
          variant="outline"
          className="w-full"
        >
          <ExternalLinkIcon className="size-4" />
          View on Mempool
        </ExternalLinkButton>
      </CardFooter>
    </Card>
  );
}

function OnchainSuccessCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-center text-base">
          Transaction Received!
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <LottieSuccess />
        <PaymentAmount amountSat={25000} />
      </CardContent>
      <CardFooter className="flex flex-col gap-3 pt-2">
        <ExternalLinkButton
          to={`https://mempool.space/tx/${DUMMY_TX_ID}`}
          variant="outline"
          className="w-full"
        >
          <ExternalLinkIcon className="size-4" />
          View on Mempool
        </ExternalLinkButton>
        <Button type="button" variant="outline" className="w-full">
          Receive Another Payment
        </Button>
        <Button variant="link" className="w-full">
          Back to Wallet
        </Button>
      </CardFooter>
    </Card>
  );
}

function BuzzPayCard() {
  return (
    <Card className="min-w-0">
      <CardHeader>
        <CardTitle className="text-base">BuzzPay PoS connection</CardTitle>
        <CardDescription>Internal app connection QR.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        <QRCode value={DUMMY_POS_URL} />
      </CardContent>
      <CardFooter className="flex flex-col gap-3">
        <Button className="w-full" variant="outline">
          <CopyIcon className="size-4" />
          Copy
        </Button>
        <Button className="w-full" variant="outline">
          <ExternalLinkIcon className="size-4" />
          Open
        </Button>
      </CardFooter>
    </Card>
  );
}

function PaymentAmount({ amountSat }: { amountSat: number }) {
  return (
    <div className="flex flex-col items-center gap-1">
      <p className="text-2xl font-medium slashed-zero">
        {new Intl.NumberFormat().format(amountSat)} sats
      </p>
      <p className="text-xl text-muted-foreground">
        {new Intl.NumberFormat("en-US", {
          currency: "USD",
          style: "currency",
        }).format((amountSat / 100_000_000) * 120000)}
      </p>
    </div>
  );
}
