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
  PayOfferRequest,
  PayOfferResponse,
  TransactionMetadata,
} from "src/types";
import { request } from "src/utils/request";

export default function Bolt12Pay() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { data: balances } = useBalances();

  const offer = state?.args?.offer as string;
  // the recipient as entered by the user (BIP-353 address or raw offer)
  const recipient = state?.args?.to as string | undefined;
  const [appId, setAppId] = React.useState<number>();
  const [amountSat, setAmountSat] = React.useState("");
  const [payerNote, setPayerNote] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const [errorMessage, setErrorMessage] = React.useState("");

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage("");
    try {
      if (!offer) {
        throw new Error("no offer set");
      }
      setLoading(true);
      const metadata: TransactionMetadata = {
        ...(recipient && { recipient_data: { identifier: recipient } }),
      };
      const payload: PayOfferRequest = {
        offer,
        amountSat: parseInt(amountSat),
        payerNote: payerNote || undefined,
        metadata,
        fromAppId: appId,
      };
      const payOfferResponse = await request<PayOfferResponse>(
        `/api/payments/offer`,
        {
          method: "POST",
          body: JSON.stringify(payload),
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      if (!payOfferResponse?.preimage) {
        throw new Error("No preimage in response");
      }
      navigate(`/wallet/send/success`, {
        state: {
          preimage: payOfferResponse.preimage,
          amountSat: parseInt(amountSat),
          to: recipient,
          pageTitle: "Send to BOLT-12 Offer",
        },
        replace: true,
      });
      toast("Successfully paid offer");
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
    if (!offer) {
      navigate("/wallet/send");
    }
  }, [navigate, offer]);

  if (!balances || !offer) {
    return <Loading />;
  }

  return (
    <div className="grid gap-4">
      <AppHeader
        pageTitle="Send to BOLT-12 Offer"
        title="Send to BOLT-12 Offer"
      />
      <div className="md:max-w-lg grid gap-4">
        <PendingPaymentAlert />
        {errorMessage && (
          <PaymentFailedAlert errorMessage={errorMessage} invoice={offer} />
        )}
        <form onSubmit={onSubmit} className="grid gap-6">
          <div className="grid gap-2">
            <div className="text-sm font-medium">Recipient</div>
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm break-all">{recipient || offer}</p>
              <Link to="/wallet/send">
                <XIcon className="w-4 h-4 cursor-pointer text-muted-foreground" />
              </Link>
            </div>
          </div>
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
          <div className="grid gap-2">
            <Label htmlFor="payerNote">Note</Label>
            <Input
              id="payerNote"
              type="text"
              value={payerNote}
              placeholder="Optional"
              onChange={(e) => {
                setPayerNote(e.target.value);
              }}
            />
          </div>
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
