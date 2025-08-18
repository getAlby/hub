import { ClipboardPasteIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { Invoice, LightningAddress } from "@getalby/lightning-tools";

export default function Send() {
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const navigate = useNavigate();

  const [recipient, setRecipient] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setRecipient(text.trim());
  };

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
      if (recipient.includes("@")) {
        const lnAddress = new LightningAddress(recipient);
        await lnAddress.fetch();
        if (lnAddress.lnurlpData) {
          navigate(`/wallet/send/lnurl-pay`, {
            state: {
              args: { lnAddress },
            },
          });
          return;
        }
      }

      const invoice = new Invoice({ pr: recipient });
      if (invoice.satoshi === 0) {
        navigate(`/wallet/send/0-amount`, {
          state: {
            args: { paymentRequest: invoice },
          },
        });
        return;
      }

      navigate(`/wallet/send/confirm-payment`, {
        state: {
          args: { paymentRequest: invoice },
        },
      });
    } catch (error) {
      toast.error("Invalid recipient", {
        description: "" + error,
      });
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  if (!balances || !channels) {
    return <Loading />;
  }

  return (
    <form onSubmit={onSubmit}>
      <div className="grid gap-2 mb-5">
        <Label htmlFor="recipient">Recipient</Label>
        <div className="flex gap-2">
          <Input
            id="recipient"
            type="text"
            value={recipient}
            autoFocus
            placeholder="Enter an invoice or Lightning Address"
            onChange={(e) => {
              setRecipient(e.target.value.trim());
            }}
          />
          <Button
            type="button"
            variant="outline"
            className="px-2"
            onClick={paste}
          >
            <ClipboardPasteIcon className="size-4" />
          </Button>
        </div>
      </div>
      <LoadingButton loading={isLoading} type="submit" disabled={!recipient}>
        Continue
      </LoadingButton>
    </form>
  );
}
