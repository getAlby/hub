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
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import LowReceivingCapacityAlert from "src/components/LowReceivingCapacityAlert";
import QRCode from "src/components/QRCode";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances";

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

  const [isLoading, setLoading] = React.useState(false);
  const [amount, setAmount] = React.useState<string>("");
  const [description, setDescription] = React.useState<string>("");
  const [transaction, setTransaction] = React.useState<Transaction | null>(
    null
  );
  const [paymentDone, setPaymentDone] = React.useState(false);
  const { data: invoiceData } = useTransaction(
    transaction ? transaction.paymentHash : "",
    true
  );

  React.useEffect(() => {
    if (invoiceData?.settledAt) {
      setPaymentDone(true);
    }
  }, [invoiceData]);

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
          amount: (parseInt(amount) || 0) * 1000,
          description,
        } as CreateInvoiceRequest),
      });

      if (invoice) {
        setTransaction(invoice);
        setAmount("");
        setDescription("");
        toast("Successfully created invoice");
      }
    } catch (e) {
      toast.error("Failed to create invoice", {
        description: "" + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const copy = () => {
    copyToClipboard(transaction?.invoice as string);
  };

  return (
    <div className="grid gap-5">
      <AppHeader title={transaction ? "Lightning Invoice" : "Create Invoice"} />
      <div className="flex flex-col md:flex-row gap-12">
        <div className="w-full md:max-w-lg grid gap-6">
          {hasChannelManagement &&
            (+amount * 1000 || transaction?.amount || 0) >=
              0.8 * balances.lightning.totalReceivable && (
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
                      <QRCode value={transaction.invoice} className="w-full" />
                      <div className="flex flex-col gap-1 items-center">
                        <p className="text-2xl font-medium slashed-zero">
                          {new Intl.NumberFormat().format(
                            Math.floor(transaction.amount / 1000)
                          )}{" "}
                          sats
                        </p>
                        <FormattedFiatAmount
                          amount={Math.floor(transaction.amount / 1000)}
                          className="text-xl"
                        />
                      </div>
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
                          {new Intl.NumberFormat().format(
                            Math.floor(transaction.amount / 1000)
                          )}{" "}
                          sats
                        </p>
                        <FormattedFiatAmount
                          amount={Math.floor(transaction.amount / 1000)}
                          className="text-xl"
                        />
                      </div>
                    </CardContent>
                    <CardFooter className="flex flex-col gap-2 pt-2">
                      <Button
                        onClick={() => {
                          setPaymentDone(false);
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
                <div className="grid gap-2">
                  <Label htmlFor="amount">Amount</Label>
                  <InputWithAdornment
                    id="amount"
                    type="number"
                    value={amount?.toString()}
                    placeholder="Amount in Satoshi..."
                    onChange={(e) => {
                      setAmount(e.target.value.trim());
                    }}
                    min={1}
                    autoFocus
                    endAdornment={
                      <FormattedFiatAmount amount={+amount} className="mr-2" />
                    }
                  />
                </div>
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
                  disabled={!amount}
                >
                  Create Invoice
                </LoadingButton>
                {(!info?.albyAccountConnected || !me?.lightning_address) && (
                  <div className="grid gap-2 border-t pt-6">
                    {!info?.albyAccountConnected &&
                      info.backendType === "LDK" && (
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
                      Receive from On-chain
                    </LinkButton>
                  </div>
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
        <ExternalLink to="https://getalby.com/auth/users/new">
          <Button variant="secondary">Get Alby Account</Button>
        </ExternalLink>
      </CardFooter>
    </Card>
  );
}
