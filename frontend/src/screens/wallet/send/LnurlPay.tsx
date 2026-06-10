import { Invoice } from "@getalby/lightning-tools";
import type { LightningAddress } from "@getalby/lightning-tools/lnurl";
import { XIcon } from "lucide-react";
import React from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { CurrencyInputField } from "src/components/CurrencyInputField";
import { InsufficientLightningBalanceAlert } from "src/components/InsufficientLightningBalanceAlert";
import Loading from "src/components/Loading";
import { PaymentFailedAlert } from "src/components/PaymentFailedAlert";
import { PendingPaymentAlert } from "src/components/PendingPaymentAlert";
import { LinkButton } from "src/components/ui/custom/link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useBalances } from "src/hooks/useBalances";
import PayFromSelect from "src/screens/wallet/send/PayFromSelect";
import {
  PayInvoiceRequest,
  PayInvoiceResponse,
  TransactionMetadata,
} from "src/types";
import { request } from "src/utils/request";

export default function LnurlPay() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { data: balances } = useBalances();

  const lnAddress = state?.args?.lnAddress as LightningAddress;
  const identifier = lnAddress.lnurlpData?.identifier;
  const [appId, setAppId] = React.useState<number>();
  const [amountSat, setAmountSat] = React.useState("");
  const [comment, setComment] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const [invoice, setInvoice] = React.useState<Invoice>();
  const [errorMessage, setErrorMessage] = React.useState("");

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage("");
    try {
      if (!lnAddress) {
        throw new Error("no lightning address set");
      }
      setLoading(true);
      const invoice = await lnAddress.requestInvoice({
        satoshi: parseInt(amountSat),
        comment,
      });
      setInvoice(invoice);
      const metadata: TransactionMetadata = {
        ...(comment && { comment }),
        ...(identifier && { recipient_data: { identifier } }),
      };
      const payload: PayInvoiceRequest = {
        metadata,
        fromAppId: appId,
      };
      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${invoice.paymentRequest}`,
        {
          method: "POST",
          body: JSON.stringify(payload),
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
        replace: true,
      });
      toast("Successfully paid invoice");
    } catch (e) {
      console.error(e);
      setErrorMessage("" + e);
      toast.error("Failed to send payment", {
        description: "" + e,
      });
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
      <AppHeader
        pageTitle="Send to Lightning Address"
        title="Send to Lightning Address"
      />
      <div className="md:max-w-lg grid gap-4">
        <PendingPaymentAlert />
        {errorMessage && invoice && (
          <PaymentFailedAlert
            errorMessage={errorMessage}
            invoice={invoice.paymentRequest}
          />
        )}
        <form onSubmit={onSubmit} className="grid gap-6">
          <div className="grid gap-2">
            <div className="text-sm font-medium">Recipient</div>
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm break-all">{lnAddress.address}</p>
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
          <CurrencyInputField
            id="amount"
            valueSat={amountSat}
            onValueSatChange={setAmountSat}
            minSat={1}
            maxSat={balances.lightning.totalSpendableSat}
            required
            autoFocus
            contextRows={[
              {
                label: "Lightning balance",
                amountSat: balances.lightning.totalSpendableSat,
              },
            ]}
          />
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
          <PayFromSelect appId={appId} onChange={setAppId} />
          <InsufficientLightningBalanceAlert amountSat={+amountSat} />
          <div className="flex gap-2">
            <LinkButton to="/wallet/send" variant="outline">
              Back
            </LinkButton>
            <LoadingButton loading={isLoading} type="submit" className="flex-1">
              Send
            </LoadingButton>
          </div>
        </form>
      </div>
    </div>
  );
}
