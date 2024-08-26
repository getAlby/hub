import { RocketIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LinkButton } from "src/components/ui/button";
import { ONCHAIN_DUST_SATS } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";

export default function BreezRedeem() {
  const { data: info } = useInfo();
  if (info?.backendType !== "BREEZ") {
    return null;
  }
  return <BreezRedeemInternal />;
}

function BreezRedeemInternal() {
  const { data: balances } = useBalances();

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
          <LinkButton to={"/wallet/withdraw"} size={"sm"}>
            Redeem
          </LinkButton>
        </AlertDescription>
      </Alert>
    </div>
  );
}
