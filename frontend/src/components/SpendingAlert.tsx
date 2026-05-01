import { AlertTriangleIcon } from "lucide-react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";

export function SpendingAlert({
  className,
  amountSat,
}: {
  className?: string;
  amountSat: number;
}) {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();

  if (!balances) {
    return null;
  }

  const maxSpendableMsat = Math.max(
    balances.lightning.nextMaxSpendableMPPMsat -
      Math.max(
        0.01 * balances.lightning.nextMaxSpendableMPPMsat,
        10000 /* fee reserve */
      ),
    0
  );

  const exceedsBalance =
    hasChannelManagement && amountSat * 1000 > maxSpendableMsat;

  if (!exceedsBalance) {
    return null;
  }

  return (
    <Alert className={className}>
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>Maximum Spendable Balance Too Low</AlertTitle>
      <AlertDescription>
        <p>
          Your payment will likely fail because your maximum spendable balance
          for the next payment is currently{" "}
          <FormattedBitcoinAmount amountMsat={maxSpendableMsat} />.
        </p>
        <div className="flex gap-2 mt-2">
          <LinkButton
            to="/channels/outgoing"
            size="sm"
            variant="secondary"
            className="w-full"
          >
            Increase Spending Balance
          </LinkButton>
        </div>
      </AlertDescription>
    </Alert>
  );
}
