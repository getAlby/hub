import { AlertTriangleIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

interface AnchorReserveAlertProps {
  amount: number;
  isSwap?: boolean;
  className?: string;
  context?: "spend" | "receive" | "swap";
}

export default function AnchorReserveAlert({
  amount,
  isSwap = false,
  className,
  context = "spend", // Default for backward compatibility
}: AnchorReserveAlertProps) {
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
    <Alert className={className} variant="default">
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>Consider your channel anchor reserves</AlertTitle>
      <AlertDescription>
        You have {channels.length} channel{channels.length > 1 ? "s" : ""} open.
        {context === "receive" ? (
          <>
            If you plan to spend from your hub wallet later, consider keeping at
            least {new Intl.NumberFormat().format(channels.length * 25000)} sats
            on-chain to maintain channel anchor reserves for potential
            force-closures.
          </>
        ) : (
          <>
            Keep at least{" "}
            {new Intl.NumberFormat().format(channels.length * 25000)} sats
            on-chain to maintain channel anchor reserves for potential
            force-closures
          </>
        )}
        {isSwap &&
          ". You can also use an external on-chain wallet to avoid this concern."}
      </AlertDescription>
    </Alert>
  );
}
