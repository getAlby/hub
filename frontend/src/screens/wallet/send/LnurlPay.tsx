import type { LightningAddress } from "@getalby/lightning-tools";
import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { TransactionMetadata } from "src/types";

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
      navigate(`/wallet/send/confirm-payment`, {
        state: {
          args: {
            paymentRequest: invoice,
            metadata,
          },
        },
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
    <form onSubmit={onSubmit} className="grid gap-4">
      <div className="grid gap-2">
        <p className="font-medium text-lg">{lnAddress.address}</p>
        {lnAddress.lnurlpData?.description && (
          <div>
            <Label>Description</Label>
            <p className="text-muted-foreground">
              {lnAddress.lnurlpData.description}
            </p>
          </div>
        )}
        <div>
          <Label htmlFor="amount">Amount</Label>
          <Input
            id="amount"
            type="number"
            value={amount}
            placeholder="Amount in Satoshi..."
            onChange={(e) => {
              setAmount(e.target.value.trim());
            }}
            min={1}
            required
            autoFocus
          />
        </div>
        {!!lnAddress.lnurlpData?.commentAllowed && (
          <div>
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
      </div>
      <div className="flex gap-4">
        <LoadingButton loading={isLoading} type="submit">
          Continue
        </LoadingButton>
        <Link to="/wallet/send">
          <Button variant="secondary">Back</Button>
        </Link>
      </div>
    </form>
  );
}
