import { AlertTriangleIcon } from "lucide-react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

export function AnchorReserveAlert({
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
    <Alert className={className} variant="warning">
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>Channel Anchor Reserves will be depleted</AlertTitle>
      <AlertDescription>
        You have channels open and by spending your entire on-chain balance
        including your anchor reserves may put your node at risk of unable to
        reclaim funds in your channel after a force-closure. To prevent this,
        set aside at least{" "}
        <FormattedBitcoinAmount amount={channels.length * 25000 * 1000} />{" "}
        on-chain.
      </AlertDescription>
    </Alert>
  );
}
