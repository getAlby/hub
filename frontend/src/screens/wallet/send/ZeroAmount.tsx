import React from "react";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import type { Invoice } from "@getalby/lightning-tools/bolt11";
import { Link, useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";

export default function ZeroAmount() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();

  const paymentRequest = state?.args?.paymentRequest as Invoice;
  const [amount, setAmount] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (!paymentRequest) {
        throw new Error("no invoice set");
      }
      setLoading(true);

      navigate(`/wallet/send/confirm-payment`, {
        state: {
          args: { paymentRequest, amount: parseInt(amount) },
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
    if (!paymentRequest) {
      navigate("/wallet/send");
    }
  }, [navigate, paymentRequest]);

  if (!paymentRequest) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-4">
      <div>
        {paymentRequest.description && (
          <div className="mt-2">
            <Label>Description</Label>
            <p className="text-muted-foreground">
              {paymentRequest.description}
            </p>
          </div>
        )}
        <div className="mb-2">
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
