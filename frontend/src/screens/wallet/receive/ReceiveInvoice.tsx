import {
  AlertTriangleIcon,
  CircleCheckIcon,
  CopyIcon,
  ReceiptTextIcon,
} from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import QRCode from "src/components/QRCode";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances";

import { useInfo } from "src/hooks/useInfo";
import { useTransaction } from "src/hooks/useTransaction";
import { copyToClipboard } from "src/lib/clipboard";
import { CreateInvoiceRequest, Transaction } from "src/types";
import { request } from "src/utils/request";

export default function ReceiveInvoice() {
  const { data: info, hasChannelManagement } = useInfo();
  const { data: me } = useAlbyMe();
  const { data: balances } = useBalances();

  const { toast } = useToast();
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
  }, [invoiceData, toast]);

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

        toast({
          title: "Successfully created invoice",
        });
      }
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to create invoice: " + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const copy = () => {
    copyToClipboard(transaction?.invoice as string, toast);
  };

  return (
    <div className="flex flex-col md:flex-row gap-6">
      <div className="w-full md:max-w-lg">
        <div className="grid gap-5">
          {hasChannelManagement &&
            parseInt(amount || "0") * 1000 >=
              0.8 * balances.lightning.totalReceivable && (
              <Alert>
                <AlertTriangleIcon className="h-4 w-4" />
                <AlertTitle>Low receiving limit</AlertTitle>
                <AlertDescription>
                  You likely won't be able to receive payments until you{" "}
                  <Link className="underline" to="/wallet/send">
                    spend
                  </Link>
                  ,{" "}
                  <Link className="underline" to="/wallet/swap?type=out">
                    swap out funds
                  </Link>
                  , or{" "}
                  <Link className="underline" to="/channels/incoming">
                    increase your receiving capacity.
                  </Link>
                </AlertDescription>
              </Alert>
            )}
          <div>
            {transaction ? (
              <Card className="w-full md:max-w-xs">
                {!paymentDone ? (
                  <>
                    <CardHeader>
                      <CardTitle className="flex justify-center">
                        <Loading className="size-4 mr-2" />
                        <p>Waiting for payment</p>
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col items-center gap-4">
                      <QRCode value={transaction.invoice} className="w-full" />
                      <div className="flex flex-col gap-2 items-center">
                        <p className="text-xl font-semibold slashed-zero">
                          {new Intl.NumberFormat().format(parseInt(amount))}{" "}
                          sats
                        </p>
                        <FormattedFiatAmount amount={parseInt(amount)} />
                      </div>
                      <div>
                        <Button onClick={copy} variant="outline">
                          <CopyIcon />
                          Copy Invoice
                        </Button>
                      </div>
                    </CardContent>
                  </>
                ) : (
                  <>
                    <CardHeader>
                      <CardTitle className="text-center">
                        Payment Received!
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col items-center gap-4">
                      <CircleCheckIcon className="w-72 h-72 p-2" />
                      <div className="flex flex-col gap-2 items-center">
                        <p className="text-xl font-semibold slashed-zero">
                          {new Intl.NumberFormat().format(parseInt(amount))}{" "}
                          sats
                        </p>
                        <FormattedFiatAmount amount={parseInt(amount)} />
                      </div>
                      <div>
                        <Button
                          variant="outline"
                          onClick={() => {
                            setPaymentDone(false);
                            setTransaction(null);
                          }}
                        >
                          Receive Another Payment
                        </Button>
                      </div>
                    </CardContent>
                  </>
                )}
              </Card>
            ) : (
              <form onSubmit={handleSubmit} className="grid gap-5">
                <div>
                  <Label htmlFor="amount">Amount</Label>
                  <Input
                    id="amount"
                    type="number"
                    value={amount?.toString()}
                    placeholder="Amount in Satoshi..."
                    onChange={(e) => {
                      setAmount(e.target.value.trim());
                    }}
                    min={1}
                    autoFocus
                  />
                  <FormattedFiatAmount amount={+amount} className="mt-2" />
                </div>
                <div>
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
                <div className="flex flex-col md:flex-row gap-4">
                  <LoadingButton
                    className="w-full md:w-auto"
                    loading={isLoading}
                    type="submit"
                    disabled={!amount}
                  >
                    Create Invoice
                  </LoadingButton>
                  {!info?.albyAccountConnected &&
                    info.backendType === "LDK" && (
                      <Link to="/wallet/receive/offer">
                        <Button variant="outline" className="w-full">
                          <ReceiptTextIcon className="h-4 w-4 shrink-0 mr-2" />
                          BOLT-12 Offer
                        </Button>
                      </Link>
                    )}
                </div>
              </form>
            )}
          </div>
        </div>
      </div>
      {!transaction &&
        (!info?.albyAccountConnected || !me?.lightning_address) && (
          <LightningAddressCard />
        )}
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
          <Button variant="secondary">Create Alby Account</Button>
        </ExternalLink>
      </CardFooter>
    </Card>
  );
}
