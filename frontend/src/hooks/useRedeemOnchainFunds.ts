import React from "react";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";

import { RedeemOnchainFundsResponse } from "src/types";
import { request } from "src/utils/request";

// TODO: move this to a different screen instead
export function useRedeemOnchainFunds() {
  const { mutate: reloadBalances } = useBalances();
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);

  const redeemFunds = React.useCallback(async () => {
    setLoading(true);
    await new Promise((resolve) => setTimeout(resolve, 100));
    const toAddress = prompt(
      "Please enter an onchain bitcoin address (bc1...) to withdraw your savings balance to another bitcoin wallet (e.g. a cold storage wallet). Make sure you own the wallet that generated this address."
    );
    if (!toAddress) {
      setLoading(false);
      return;
    }

    const confirmAddress = prompt(
      "Please confirm an onchain bitcoin address. Make sure you own the wallet that generated this address!"
    );
    if (toAddress !== confirmAddress) {
      setLoading(false);
      return;
    }

    if (
      !confirm(
        "Are you sure you want to send your onchain funds to another wallet? if you send to an address you do not own, your funds will be lost."
      )
    ) {
      setLoading(false);
      return;
    }

    try {
      const response = await request<RedeemOnchainFundsResponse>(
        "/api/wallet/redeem-onchain-funds",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ toAddress }),
        }
      );
      console.info("Redeemed onchain funds", response);
      if (!response?.txId) {
        throw new Error("No address in response");
      }
      prompt("Funds redeemed. Copy TX to view in mempool", response.txId);
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to request a new address",
        description: "" + error,
      });
    } finally {
      setLoading(false);
    }

    await reloadBalances();
  }, [reloadBalances, toast]);

  return React.useMemo(
    () => ({ redeemFunds, isLoading }),
    [isLoading, redeemFunds]
  );
}
