import { AlertTriangleIcon } from "lucide-react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";

export function InsufficientLightningBalanceAlert({
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
      <AlertTitle>Not enough spendable balance</AlertTitle>
      <AlertDescription>
        <p>
          This payment is above the wallet's current spendable balance. The most
          you can send right now is{" "}
          <FormattedBitcoinAmount amountMsat={maxSpendableMsat} />.
        </p>
        <div className="flex gap-2 mt-2 items-center justify-center">
          <LinkButton
            to="/wallet/swap?type=in"
            size="sm"
            variant="secondary"
            className="flex-1"
          >
            Swap In
          </LinkButton>
          <p>or</p>
          <LinkButton
            to="/channels/outgoing"
            size="sm"
            variant="secondary"
            className="flex-1"
          >
            Open Outbound Channel
          </LinkButton>
        </div>
      </AlertDescription>
    </Alert>
  );
}
