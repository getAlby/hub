import { Invoice } from "@getalby/lightning-tools/bolt11";
import { LightningAddress } from "@getalby/lightning-tools/lnurl";
import { validate as validateBitcoinAddress } from "bitcoin-address-validation";
import { ClipboardPasteIcon } from "lucide-react";
import React from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { CryptoSwapAlert } from "src/components/CryptoSwapAlert";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { parseBip21 } from "src/utils/parseBip21";

export default function Send() {
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [recipient, setRecipient] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const [showSwapAlert, setShowSwapAlert] = React.useState(false);

  React.useEffect(() => {
    const uri = searchParams.get("bip21");
    if (!uri) {
      return;
    }
    try {
      const bip21 = parseBip21(uri);
      if (bip21.lightning) {
        const invoice = new Invoice({ pr: bip21.lightning });
        if (invoice.satoshi === 0) {
          navigate(`/wallet/send/0-amount`, {
            state: { args: { paymentRequest: invoice } },
            replace: true,
          });
        } else {
          navigate(`/wallet/send/confirm-payment`, {
            state: { args: { paymentRequest: invoice } },
            replace: true,
          });
        }
        return;
      }
      if (bip21.address) {
        if (!validateBitcoinAddress(bip21.address)) {
          throw new Error("invalid bitcoin address");
        }
        navigate(`/wallet/send/onchain`, {
          state: {
            args: {
              address: bip21.address,
              amount: bip21.amount ? String(bip21.amount) : undefined,
            },
          },
          replace: true,
        });
      }
    } catch (error) {
      toast.error("Invalid Bitcoin URI", { description: "" + error });
    }
  }, [searchParams, navigate]);

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setRecipient(text.trim());
  };

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      setLoading(true);
      if (validateBitcoinAddress(recipient)) {
        navigate(`/wallet/send/onchain`, {
          state: {
            args: { address: recipient },
          },
        });
        return;
      }

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
      setShowSwapAlert(true);
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
    <div className="grid gap-4">
      <AppHeader pageTitle="Send" title="Send" />
      <div className="w-full md:max-w-lg">
        <form onSubmit={onSubmit} className="grid gap-6">
          {showSwapAlert && <CryptoSwapAlert />}
          <div className="grid gap-2">
            <Label htmlFor="recipient">Recipient</Label>
            <div className="flex gap-2">
              <Input
                id="recipient"
                type="text"
                value={recipient}
                autoFocus
                placeholder="Invoice, lightning address, on-chain address"
                onChange={(e) => {
                  setRecipient(e.target.value.trim());
                  setShowSwapAlert(false);
                }}
              />
              <Button
                type="button"
                variant="outline"
                className="px-2"
                onClick={paste}
              >
                <ClipboardPasteIcon className="w-4 h-4" />
              </Button>
            </div>
          </div>
          <LoadingButton
            loading={isLoading}
            type="submit"
            disabled={!recipient}
            className="flex-1"
          >
            Continue
          </LoadingButton>
        </form>
      </div>
    </div>
  );
}
