import type { LightningAddress } from "@getalby/lightning-tools/lnurl";
import { XIcon } from "lucide-react";
import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import { PendingPaymentAlert } from "src/components/PendingPaymentAlert";
import { SpendingAlert } from "src/components/SpendingAlert";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { PayInvoiceResponse, TransactionMetadata } from "src/types";
import { request } from "src/utils/request";

export default function LnurlPay() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { data: balances } = useBalances();

  const lnAddress = state?.args?.lnAddress as LightningAddress;
  const identifier = lnAddress.lnurlpData?.identifier;
  const [amount, setAmount] = React.useState("");
  const [comment, setComment] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (!lnAddress) {
        throw new Error("no lightning address set");
      }
      setLoading(true);
      const invoice = await lnAddress.requestInvoice({
        satoshi: parseInt(amount),
        comment,
      });
      const metadata: TransactionMetadata = {
        ...(comment && { comment }),
        ...(identifier && { recipient_data: { identifier } }),
      };
      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${invoice.paymentRequest}`,
        {
          method: "POST",
          body: JSON.stringify({
            metadata,
          }),
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      if (!payInvoiceResponse?.preimage) {
        throw new Error("No preimage in response");
      }
      navigate(`/wallet/send/success`, {
        state: {
          preimage: payInvoiceResponse.preimage,
          invoice,
          to: lnAddress.address,
          pageTitle: "Send to Lightning Address",
        },
      });
      toast({
        title: "Successfully paid invoice",
      });
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to send payment",
        description: "" + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    if (!lnAddress) {
      navigate("/wallet/send");
    }
  }, [navigate, lnAddress]);

  if (!balances || !lnAddress) {
    return <Loading />;
  }

  return (
    <div className="grid gap-4">
      <AppHeader title="Send to Lightning Address" />
      <PendingPaymentAlert />
      <form onSubmit={onSubmit} className="grid gap-6 md:max-w-lg">
        <div className="grid gap-2">
          <div className="text-sm font-medium">Recipient</div>
          <div className="flex items-center justify-between">
            <p className="text-sm">{lnAddress.address}</p>
            <Link to="/wallet/send">
              <XIcon className="w-4 h-4 cursor-pointer text-muted-foreground" />
            </Link>
          </div>
        </div>
        {lnAddress.lnurlpData?.description && (
          <div className="grid gap-2">
            <Label>Description</Label>
            <p className="text-muted-foreground text-sm">
              {lnAddress.lnurlpData.description}
            </p>
          </div>
        )}
        <div className="grid gap-2">
          <Label htmlFor="amount">Amount</Label>
          <InputWithAdornment
            id="amount"
            type="number"
            value={amount}
            placeholder="Amount in Satoshi..."
            onChange={(e) => {
              setAmount(e.target.value.trim());
            }}
            min={1}
            max={Math.floor(balances.lightning.totalSpendable / 1000)}
            required
            autoFocus
            className="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
            endAdornment={
              <FormattedFiatAmount amount={Number(amount)} className="mr-2" />
            }
          />
          <div className="grid gap-2">
            <div className="flex justify-between text-xs text-muted-foreground sensitive slashed-zero">
              <div>
                Spending Balance:{" "}
                {new Intl.NumberFormat().format(
                  Math.floor(balances.lightning.totalSpendable / 1000)
                )}{" "}
                sats
              </div>
              <FormattedFiatAmount
                className="text-xs"
                amount={Math.floor(balances.lightning.totalSpendable / 1000)}
              />
            </div>
          </div>
        </div>
        {!!lnAddress.lnurlpData?.commentAllowed && (
          <div className="grid gap-2">
            <Label htmlFor="comment">Comment</Label>
            <Input
              id="comment"
              type="text"
              value={comment}
              placeholder="Optional"
              onChange={(e) => {
                setComment(e.target.value);
              }}
            />
          </div>
        )}
        <SpendingAlert amount={+amount} />
        <div className="flex gap-2">
          <LinkButton to="/wallet/send" variant="outline">
            Back
          </LinkButton>
          <LoadingButton
            loading={isLoading}
            type="submit"
            className="w-full md:w-fit"
          >
            Send
          </LoadingButton>
        </div>
      </form>
    </div>
  );
}
