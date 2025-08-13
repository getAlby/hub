import { AlertTriangleIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";

export function SpendingAlert({
  className,
  amount,
}: {
  className?: string;
  amount: number;
}) {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();

  if (!balances) {
    return null;
  }

  const maxSpendable = Math.max(
    balances.lightning.nextMaxSpendableMPP -
      Math.max(
        0.01 * balances.lightning.nextMaxSpendableMPP,
        10000 /* fee reserve */
      ),
    0
  );

  const exceedsBalance = hasChannelManagement && amount * 1000 > maxSpendable;

  if (!exceedsBalance) {
    return null;
  }

  return (
    <Alert className={className}>
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>Maximum Spendable Balance Too Low</AlertTitle>
      <AlertDescription>
        Your payment will likely fail because your maximum spendable balance for
        the next payment is currently{" "}
        {new Intl.NumberFormat().format(Math.floor(maxSpendable / 1000))} sats.
      </AlertDescription>
    </Alert>
  );
}
