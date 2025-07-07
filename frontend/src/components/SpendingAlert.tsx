import { AlertTriangleIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";

export function SpendingAlert({
  className,
  maxSpendable,
}: {
  className?: string;
  maxSpendable: number;
}) {
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
