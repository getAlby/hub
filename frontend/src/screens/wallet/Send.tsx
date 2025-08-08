import { ClipboardPasteIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { Invoice } from "@getalby/lightning-tools/bolt11";
import { LightningAddress } from "@getalby/lightning-tools/lnurl";

export default function Send() {
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { toast } = useToast();
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
      toast({
        variant: "destructive",
        title: "Invalid recipient",
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
      <Label htmlFor="recipient">Recipient</Label>
      <div className="flex gap-2 mb-4">
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
      <LoadingButton loading={isLoading} type="submit" disabled={!recipient}>
        Continue
      </LoadingButton>
    </form>
  );
}
