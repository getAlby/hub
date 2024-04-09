import { RocketIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LoadingButton } from "src/components/ui/loading-button";
import { ONCHAIN_DUST_SATS } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useRedeemOnchainFunds } from "src/hooks/useRedeemOnchainFunds";

export default function BreezRedeem() {
  const { data: info } = useInfo();
  if (info?.backendType !== "BREEZ") {
    return null;
  }
  return <BreezRedeemInternal />;
}

function BreezRedeemInternal() {
  const { data: balances } = useBalances();

  const redeemOnchainFunds = useRedeemOnchainFunds();

  if (!balances || balances.onchain.spendable <= ONCHAIN_DUST_SATS) {
    return null;
  }

  return (
    <div className="mb-8">
      <Alert>
        <RocketIcon className="h-4 w-4" />
        <AlertTitle>Breez channel closed</AlertTitle>
        <AlertDescription>
          <div className="mb-2">
            One of your Breez channels was closed and you have{" "}
            {balances.onchain.spendable} sats to redeem.
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
