import React from "react";
import Alert from "src/components/Alert";
import Loading from "src/components/Loading";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { useOnchainBalance } from "src/hooks/useOnchainBalance";
import { RedeemOnchainFundsResponse } from "src/types";
import { request } from "src/utils/request";

export default function BreezRedeem() {
  const { data: info } = useInfo();
  if (info?.backendType !== "BREEZ") {
    return null;
  }
  return <BreezRedeemInternal />;
}

function BreezRedeemInternal() {
  const { data: onchainBalance, mutate: reloadOnchainBalance } =
    useOnchainBalance();

  const { data: csrf } = useCSRF();
  const [isLoading, setLoading] = React.useState(false);

  const redeemFunds = React.useCallback(async () => {
    if (!csrf) {
      return;
    }
    const toAddress = prompt(
      "Please enter an onchain bitcoin address (bc1...)"
    );
    if (!toAddress) {
      return;
    }
    setLoading(true);

    try {
      const response = await request<RedeemOnchainFundsResponse>(
        "/api/wallet/redeem-onchain-funds",
        {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ toAddress }),
        }
      );
      console.log("Redeemed onchain funds", response);
      if (!response?.txId) {
        throw new Error("No address in response");
      }
      prompt("Funds redeemed. Copy TX to view in mempool", response.txId);
    } catch (error) {
      alert("Failed to request a new address: " + error);
    } finally {
      setLoading(false);
    }

    await reloadOnchainBalance();
  }, [csrf, reloadOnchainBalance]);

  if (!onchainBalance || onchainBalance.spendable <= 0) {
    return null;
  }

  return (
    <div className="mb-8">
      <Alert type="info">
        <div className="flex justify-between items-center text-sm">
          One of your Breez channels was closed and you have{" "}
          {onchainBalance.spendable} sats to redeem.{" "}
          <button
            className="flex justify-center items-center gap-2 bg-purple-100 p-2 text-purple-500 rounded-md"
            onClick={redeemFunds}
            disabled={isLoading}
          >
            Redeem onchain funds {isLoading && <Loading />}
          </button>
        </div>
      </Alert>
    </div>
  );
}
