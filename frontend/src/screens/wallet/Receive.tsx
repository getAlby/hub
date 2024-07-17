import confetti from "canvas-confetti";
import { AlertTriangle, ArrowDown, CircleCheck, CopyIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
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
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { useTransaction } from "src/hooks/useTransaction";
import { copyToClipboard } from "src/lib/clipboard";
import { CreateInvoiceRequest, Transaction } from "src/types";
import { request } from "src/utils/request";

export default function Receive() {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);
  const [amount, setAmount] = React.useState<string>("");
  const [description, setDescription] = React.useState<string>("");
  const [invoice, setInvoice] = React.useState<Transaction | null>(null);
  const [paymentDone, setPaymentDone] = React.useState(false);
  const { data: invoiceData } = useTransaction(
    invoice ? invoice.payment_hash : "",
    true
  );

  React.useEffect(() => {
    if (invoiceData?.settled_at) {
      setPaymentDone(true);
      popConfetti();
      toast({
        title: "Payment received!",
      });
    }
  }, [invoiceData, toast]);

  if (!balances || !channels) {
    return <Loading />;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!csrf) {
      throw new Error("csrf not loaded");
    }
    try {
      setLoading(true);
      const invoice = await request<Transaction>("/api/invoices", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          amount: (parseInt(amount) || 0) * 1000,
          description,
        } as CreateInvoiceRequest),
      });
      setAmount("");
      setDescription("");
      if (invoice) {
        setInvoice(invoice);

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
    copyToClipboard(invoice?.invoice as string);
    toast({ title: "Copied to clipboard." });
  };

  const popConfetti = () => {
    for (let i = 0; i < 10; i++) {
      setTimeout(
        () => {
          confetti({
            origin: {
              x: Math.random(),
              y: Math.random(),
            },
            colors: ["#000", "#333", "#666", "#999", "#BBB", "#FFF"],
          });
        },
        Math.floor(Math.random() * 1000)
      );
    }
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Receive"
        description="Create a lightning invoice that can be paid by any bitcoin lightning wallet"
      />
      {!!channels?.length && (
        <>
          {/* If all channels have less than 20% incoming capacity, show a warning */}
          {channels?.every(
            (channel) =>
              channel.remoteBalance <
                (channel.localBalance + channel.remoteBalance) * 0.2 ||
              parseInt(amount) * 1000 > channel.remoteBalance
          ) && (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Low receiving capacity</AlertTitle>
              <AlertDescription>
                You likely won't be able to receive payments until you{" "}
                <Link className="underline" to="/channels/incoming">
                  increase your receiving capacity.
                </Link>
              </AlertDescription>
            </Alert>
          )}
        </>
      )}
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          {invoice ? (
            <>
              <Card className="w-full">
                <CardHeader>
                  <CardTitle className="text-center">
                    {paymentDone ? "Payment Received" : "Invoice"}
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex flex-col items-center gap-4">
                  {paymentDone ? (
                    <>
                      <CircleCheck className="w-32 h-32 mb-2" />
                      <p>Received {(invoiceData?.amount ?? 0) / 1000} sats</p>
                    </>
                  ) : (
                    <>
                      <div className="flex flex-row items-center gap-2 text-sm">
                        <Loading className="w-4 h-4" />
                        <p>Waiting for payment</p>
                      </div>
                      <QRCode value={invoice.invoice} className="w-full" />
                      <div>
                        <Button onClick={copy} variant="outline">
                          <CopyIcon className="w-4 h-4 mr-2" />
                          Copy Invoice
                        </Button>
                      </div>
                    </>
                  )}
                </CardContent>
              </Card>
              {paymentDone && (
                <>
                  <Button
                    className="mt-4 w-full"
                    onClick={() => {
                      setPaymentDone(false);
                      setInvoice(null);
                    }}
                  >
                    Receive Another Payment
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
              <div>
                <LoadingButton
                  loading={isLoading}
                  type="submit"
                  disabled={!amount}
                >
                  Create Invoice
                </LoadingButton>
              </div>
            </form>
          )}
        </div>
        {hasChannelManagement && (
          <Card className="w-full hidden md:block self-start">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Receiving Capacity
              </CardTitle>
              <ArrowDown className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {!balances && (
                <div>
                  <div className="animate-pulse d-inline ">
                    <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
                  </div>
                </div>
              )}
              <div className="text-2xl font-bold">
                {balances && (
                  <>
                    {new Intl.NumberFormat().format(
                      Math.floor(balances.lightning.totalReceivable / 1000)
                    )}{" "}
                    sats
                  </>
                )}
              </div>
            </CardContent>

            <CardFooter className="flex justify-end">
              <Link to="/channels/incoming">
                <Button variant="outline">Increase</Button>
              </Link>
            </CardFooter>
          </Card>
        )}
      </div>
    </div>
  );
}
