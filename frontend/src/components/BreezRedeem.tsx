import Alert from "src/components/Alert";
import Loading from "src/components/Loading";
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
      <Alert type="info">
        <div className="flex justify-between items-center text-sm">
          One of your Breez channels was closed and you have{" "}
          {balances.onchain.spendable} sats to redeem.{" "}
          <button
            className="flex justify-center items-center gap-2 bg-purple-100 p-2 text-purple-500 rounded-md"
            onClick={redeemOnchainFunds.redeemFunds}
            disabled={redeemOnchainFunds.isLoading}
          >
            Redeem onchain funds {redeemOnchainFunds.isLoading && <Loading />}
          </button>
        </div>
      </Alert>
    </div>
  );
}
