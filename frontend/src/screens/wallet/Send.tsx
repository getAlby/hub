import { ClipboardPaste } from "lucide-react";
import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { Invoice, LightningAddress } from "@getalby/lightning-tools";

export default function Send() {
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { toast } = useToast();
  const navigate = useNavigate();
  const { search } = useLocation();

  const queryParams = new URLSearchParams(search);
  const recipientParam = queryParams.get("recipient") ?? "";

  const [recipient, setRecipient] = React.useState<string>(recipientParam);
  const [isLoading, setLoading] = React.useState(false);

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setRecipient(text.trim());
  };

  const handleNavigation = React.useCallback(
    async (recipient: string) => {
      try {
        setLoading(true);
        const lnAddress = new LightningAddress(recipient);
        await lnAddress.fetch();
        if (lnAddress.lnurlpData) {
          navigate(`/wallet/send/lnurl-pay`, {
            state: {
              args: { lnurlDetails: lnAddress.lnurlpData },
            },
          });
          return;
        }

        const invoice = new Invoice({ pr: recipient });
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
    },
    [navigate, toast]
  );

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await handleNavigation(recipient);
  };

  React.useEffect(() => {
    if (recipientParam) {
      handleNavigation(recipientParam);
    }
  }, [handleNavigation, recipientParam]);

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
