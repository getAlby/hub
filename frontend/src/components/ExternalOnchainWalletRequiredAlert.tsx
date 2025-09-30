import { AlertTriangleIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

export function ExternalOnchainWalletRequiredAlert({
  amount,
  className,
}: {
  amount: number;
  className?: string;
}) {
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();

  if (!balances || !channels) {
    return null;
  }

  const showAlert =
    amount &&
    !!channels.length &&
    +amount > balances.onchain.spendable - channels.length * 25000;

  if (!showAlert) {
    return null;
  }

  return (
    <Alert className={className}>
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>External Wallet Required</AlertTitle>
      <AlertDescription>
        It's recommended to make this payment with an external on-chain wallet
        as you do not have enough funds in your hub's on-chain wallet to make
        this payment. To protect your anchor reserves, you need to set aside at
        least {new Intl.NumberFormat().format(channels.length * 25000)} sats
        on-chain.
      </AlertDescription>
    </Alert>
  );
}
