import { Invoice } from "@getalby/lightning-tools";
import {
  AlertTriangle,
  ArrowUp,
  CircleCheck,
  ClipboardPaste,
  CopyIcon,
} from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
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
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";
import { copyToClipboard } from "src/lib/clipboard";
import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

export default function Send() {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();

  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);
  const [invoice, setInvoice] = React.useState("");
  const [invoiceDetails, setInvoiceDetails] = React.useState<Invoice | null>(
    null
  );
  const [payResponse, setPayResponse] =
    React.useState<PayInvoiceResponse | null>(null);
  const [paymentDone, setPaymentDone] = React.useState(false);

  if (!balances || !channels) {
    return <Loading />;
  }

  const handleContinue = () => {
    try {
      setInvoiceDetails(new Invoice({ pr: invoice }));
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Invalid Payment Request",
      });
      console.error(error);
      setInvoice("");
    }
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${invoice}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      if (payInvoiceResponse) {
        setPayResponse(payInvoiceResponse);
        setPaymentDone(true);
        setInvoice("");
        toast({
          title: "Successfully paid invoice",
        });
      }
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to send: " + e,
      });
      setInvoice("");
      setInvoiceDetails(null);
      console.error(e);
    }
    setLoading(false);
  };

  const copy = () => {
    copyToClipboard(payResponse?.preimage as string);
    toast({ title: "Copied to clipboard." });
  };

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setInvoice(text.trim());
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Send"
        description="Pay a lightning invoice created by any bitcoin lightning wallet"
      />
      {hasChannelManagement &&
        (invoiceDetails?.satoshi || 0) * 1000 >=
          0.8 * balances.lightning.totalSpendable && (
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Low spending balance</AlertTitle>
            <AlertDescription>
              You won't be able to make payments until you{" "}
              <Link className="underline" to="/channels/outgoing">
                increase your spending balance.
              </Link>
            </AlertDescription>
          </Alert>
        )}
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          {paymentDone ? (
            <>
              <Card className="w-full">
                <CardHeader>
                  <CardTitle className="text-center">
                    Payment Successful
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex flex-col items-center gap-4">
                  <CircleCheck className="w-32 h-32 mb-2" />
                  <Button onClick={copy} variant="outline">
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Copy Preimage
                  </Button>
                </CardContent>
              </Card>
              {paymentDone && (
                <>
                  <Button
                    className="mt-4 w-full"
                    onClick={() => {
                      setPaymentDone(false);
                      setInvoiceDetails(null);
                      setPayResponse(null);
                      setInvoice("");
                    }}
                  >
                    Make Another Payment
                  </Button>
                  <Link to="/wallet">
                    <Button
                      className="mt-4 w-full"
                      onClick={() => {
                        setPaymentDone(false);
                      }}
                      variant="secondary"
                    >
                      Back To Wallet
                    </Button>
                  </Link>
                </>
              )}
            </>
          ) : invoiceDetails ? (
            <form onSubmit={handleSubmit} className="grid gap-5">
              <div className="">
                <p className="text-lg mb-5">Payment Details</p>
                <p className="font-bold">{invoiceDetails.satoshi} sats</p>
                <p className="text-muted-foreground">
                  {invoiceDetails.description}
                </p>
              </div>
              <div className="flex gap-5">
                <LoadingButton
                  loading={isLoading}
                  type="submit"
                  disabled={!invoice}
                  autoFocus
                >
                  Confirm Payment
                </LoadingButton>
                <Button
                  onClick={() => setInvoiceDetails(null)}
                  variant="secondary"
                >
                  Back
                </Button>
              </div>
            </form>
          ) : (
            <div className="grid gap-5">
              <div className="">
                <Label htmlFor="recipient">Recipient</Label>
                <div className="flex gap-2">
                  <Input
                    id="recipient"
                    type="text"
                    value={invoice}
                    autoFocus
                    placeholder="Enter an invoice"
                    onChange={(e) => {
                      setInvoice(e.target.value.trim());
                    }}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    className="px-2"
                    onClick={paste}
                  >
                    <ClipboardPaste className="w-4 h-4" />
                  </Button>
                </div>
              </div>
              <div>
                <Button onClick={handleContinue} disabled={!invoice}>
                  Continue
                </Button>
              </div>
            </div>
          )}
        </div>
        <Card className="w-full hidden md:block self-start">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Spending Balance
            </CardTitle>
            <ArrowUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {!balances && (
              <div>
                <div className="animate-pulse d-inline ">
                  <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                </div>
              </div>
            )}
            {balances && (
              <div className="text-2xl font-bold balance sensitive ph-no-capture">
                {new Intl.NumberFormat(undefined, {}).format(
                  Math.floor(balances.lightning.totalSpendable / 1000)
                )}{" "}
                sats
              </div>
            )}
          </CardContent>
          {hasChannelManagement && (
            <CardFooter className="flex justify-end">
              <Link to="/channels/outgoing">
                <Button variant="outline">Top Up</Button>
              </Link>
            </CardFooter>
          )}
        </Card>
      </div>
    </div>
  );
}
