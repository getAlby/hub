import { RocketIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LoadingButton } from "src/components/ui/loading-button";
import { useInfo } from "src/hooks/useInfo";
import { useOnchainBalance } from "src/hooks/useOnchainBalance";
import { useRedeemOnchainFunds } from "src/hooks/useRedeemOnchainFunds";

export default function BreezRedeem() {
  const { data: info } = useInfo();
  if (info?.backendType !== "BREEZ") {
    return null;
  }
  return <BreezRedeemInternal />;
}

function BreezRedeemInternal() {
  const { data: onchainBalance } = useOnchainBalance();

  const redeemOnchainFunds = useRedeemOnchainFunds();

  if (!onchainBalance || onchainBalance.spendable <= 0) {
    return null;
  }

  return (
    <div className="mb-8">
      <Alert>
        <RocketIcon className="h-4 w-4" />
        <AlertTitle>Breez channel closed</AlertTitle>
        <AlertDescription>
          <div className="mb-2">
            One of your Breez channels was closed and you have {1} sats to
            redeem.
          </div>
          <LoadingButton
            size={"sm"}
            loading={redeemOnchainFunds.isLoading}
            onClick={redeemOnchainFunds.redeemFunds}
            disabled={redeemOnchainFunds.isLoading}
          >
            Redeem
          </LoadingButton>
        </AlertDescription>
      </Alert>
    </div>
  );
}
