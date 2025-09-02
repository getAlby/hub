import { AlertTriangleIcon } from "lucide-react";
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
    !!channels.length &&
    +amount > balances.onchain.spendable - channels.length * 25000;

  if (!showAlert) {
    return null;
  }

  return (
    <Alert className={className}>
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>Channel Anchor Reserves may be depleted</AlertTitle>
      <AlertDescription>
        You have channels open and this withdrawal may deplete your anchor
        reserves which may make it harder to close channels without depositing
        additional onchain funds to your savings balance. To avoid this, set
        aside at least {new Intl.NumberFormat().format(channels.length * 25000)}{" "}
        sats on-chain.
      </AlertDescription>
    </Alert>
  );
}
