import { Invoice, LightningAddress } from "@getalby/lightning-tools";
import { AlertTriangle, ClipboardPaste } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import BalanceCard from "src/components/wallet/BalanceCard";
import ConfirmPayment from "src/components/wallet/ConfirmPayment";
import LnurlPay from "src/components/wallet/LnurlPay";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";

// email regex: https://emailregex.com/
// modified to allow _ in subdomains
const LIGHTNING_ADDRESS_REGEX =
  /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-_0-9]+\.)+[a-zA-Z]{2,}))$/;

const BOLT11_REGEX = /^(lnbc|lntbs)([0-9a-z]+)$/i;

export default function Send() {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();
  const { toast } = useToast();

  const [amount, setAmount] = React.useState<number>(0);
  const [recipient, setRecipient] = React.useState<string>("");
  const [lnurlDetails, setLnurlDetails] = React.useState<
    Invoice | LightningAddress
  >();

  const paste = async () => {
    const text = await navigator.clipboard.readText();
    setRecipient(text.trim());
  };

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (LIGHTNING_ADDRESS_REGEX.test(recipient)) {
        const lnAddress = new LightningAddress(recipient);
        await lnAddress.fetch();
        if (!lnAddress.lnurlpData) {
          throw new Error("invalid lightning address");
        }
        setLnurlDetails(lnAddress);
      } else if (BOLT11_REGEX.test(recipient)) {
        const invoice = new Invoice({ pr: recipient });
        setAmount(invoice.satoshi);
        setLnurlDetails(invoice);
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
    }
  };

  const onReset = async () => {
    setAmount(0);
    setRecipient("");
    setLnurlDetails(undefined);
  };

  if (!balances || !channels) {
    return <Loading />;
  }

  const hasLowBalance =
    hasChannelManagement &&
    (amount || 0) * 1000 >= 0.8 * balances.lightning.totalSpendable;

  const renderScreen = () => {
    if (lnurlDetails instanceof LightningAddress) {
      return (
        <LnurlPay
          lnAddress={lnurlDetails}
          onReset={onReset}
          onAmountChange={setAmount}
        />
      );
    } else if (lnurlDetails instanceof Invoice) {
      return <ConfirmPayment invoice={lnurlDetails} onReset={onReset} />;
    } else {
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
          <Button type="submit" disabled={!recipient}>
            Continue
          </Button>
        </form>
      );
    }
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Send"
        description="Pay a lightning invoice created by any bitcoin lightning wallet"
      />
      {hasLowBalance && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Low spending balance</AlertTitle>
          <AlertDescription>
            You won't be able to make payments until you{" "}
            <Link className="underline" to="/channels/outgoing">
              increase your spending balance.
            </Link>
          </AlertDescription>
        </Alert>
      )}
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">{renderScreen()}</div>
        <BalanceCard
          balances={balances}
          hasChannelManagement={!!hasChannelManagement}
          isSpending
        />
      </div>
    </div>
  );
}
