import { AlertTriangleIcon, CircleCheckIcon, CopyIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
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
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";

import { useInfo } from "src/hooks/useInfo";
import { useTransaction } from "src/hooks/useTransaction";
import { copyToClipboard } from "src/lib/clipboard";
import { CreateInvoiceRequest, Transaction } from "src/types";
import { request } from "src/utils/request";

export default function ReceiveInvoice() {
  const { hasChannelManagement } = useInfo();
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

  if (!balances) {
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
    <div className="grid gap-5">
      {hasChannelManagement &&
        parseInt(amount || "0") * 1000 >=
          0.8 * balances.lightning.totalReceivable && (
          <Alert>
            <AlertTriangleIcon className="h-4 w-4" />
            <AlertTitle>Low receiving capacity</AlertTitle>
            <AlertDescription>
              You likely won't be able to receive payments until you{" "}
              <Link className="underline" to="/channels/incoming">
                increase your receiving capacity.
              </Link>
            </AlertDescription>
          </Alert>
        )}
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          {transaction ? (
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
                      <CircleCheckIcon className="w-32 h-32 mb-1" />
                      <div className="flex flex-col gap-2 items-center">
                        <p>
                          Received{" "}
                          {Math.floor((invoiceData?.amount ?? 0) / 1000)} sats
                        </p>
                        <FormattedFiatAmount
                          amount={Math.floor((invoiceData?.amount ?? 0) / 1000)}
                        />
                      </div>
                    </>
                  ) : (
                    <>
                      <div className="flex flex-col gap-2 items-center">
                        <p className="text-xl slashed-zero">
                          {new Intl.NumberFormat().format(parseInt(amount))}{" "}
                          sats
                        </p>
                        <FormattedFiatAmount amount={parseInt(amount)} />
                      </div>
                      <div className="flex flex-row items-center gap-2 text-sm">
                        <Loading className="w-4 h-4" />
                        <p>Waiting for payment</p>
                      </div>
                      <QRCode value={transaction.invoice} className="w-full" />
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
                      setTransaction(null);
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
      </div>
    </div>
  );
}
