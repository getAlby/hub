import React from "react";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";

export function useSyncWallet() {
  const { data: csrf } = useCSRF();
  const REQUEST_WALLET_SYNC_INTERVAL = 30_000; // request a wallet sync every 30s (NOTE: it won't actually sync this often)

  React.useEffect(() => {
    if (!csrf) {
      return;
    }

    const intervalId = setInterval(async () => {
      try {
        await request("/api/wallet/sync", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
        });
      } catch (error) {
        console.error("failed to request wallet sync", error);
      }
    }, REQUEST_WALLET_SYNC_INTERVAL);

    return () => {
      clearInterval(intervalId);
    };
  }, [csrf]);

  return null;
}
