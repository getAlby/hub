import React from "react";
import { useBalances } from "src/hooks/useBalances";
import { useCSRF } from "src/hooks/useCSRF";
import { RedeemOnchainFundsResponse } from "src/types";
import { request } from "src/utils/request";

export function useRedeemOnchainFunds() {
  const { data: csrf } = useCSRF();
  const { mutate: reloadBalances } = useBalances();
  const [isLoading, setLoading] = React.useState(false);

  const redeemFunds = React.useCallback(async () => {
    if (!csrf) {
      return;
    }
    setLoading(true);
    await new Promise((resolve) => setTimeout(resolve, 100));
    const toAddress = prompt(
      "Please enter an onchain bitcoin address (bc1...)"
    );
    if (!toAddress) {
      setLoading(false);
      return;
    }

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

    await reloadBalances();
  }, [csrf, reloadBalances]);

  return React.useMemo(
    () => ({ redeemFunds, isLoading }),
    [isLoading, redeemFunds]
  );
}
