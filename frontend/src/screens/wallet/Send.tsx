import { ClipboardPaste } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { Invoice, LightningAddress } from "@getalby/lightning-tools";

// email regex: https://emailregex.com/
// modified to allow _ in subdomains
const LIGHTNING_ADDRESS_REGEX =
  /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-_0-9]+\.)+[a-zA-Z]{2,}))$/;

const BOLT11_REGEX = /^(lnbc|lntbs)([0-9a-z]+)$/i;

export default function Send() {
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { toast } = useToast();
  const navigate = useNavigate();

  const [recipient, setRecipient] = React.useState<string>("");
  const [isLoading, setLoading] = React.useState(false);

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setRecipient(text.trim());
  };

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
      if (LIGHTNING_ADDRESS_REGEX.test(recipient)) {
        const lnAddress = new LightningAddress(recipient);
        await lnAddress.fetch();
        if (!lnAddress.lnurlpData) {
          throw new Error("invalid lightning address");
        }
        navigate(`/wallet/send/lnurl-pay`, {
          state: {
            args: { lnAddress: lnAddress },
          },
        });
      } else if (BOLT11_REGEX.test(recipient)) {
        const invoice = new Invoice({ pr: recipient });
        navigate(`/wallet/send/confirm-payment`, {
          state: {
            args: { paymentRequest: invoice },
          },
        });
      } else {
        throw new Error("invalid recipient");
      }
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to send payment",
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
          <ClipboardPaste className="w-4 h-4" />
        </Button>
      </div>
      <LoadingButton loading={isLoading} type="submit" disabled={!recipient}>
        Continue
      </LoadingButton>
    </form>
  );
}
