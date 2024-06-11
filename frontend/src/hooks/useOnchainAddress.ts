import useSWR from "swr";

import React from "react";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";
import { swrFetcher } from "src/utils/swr";

export function useOnchainAddress() {
  const { data: csrf } = useCSRF();
  const swr = useSWR<string>("/api/wallet/address", swrFetcher);
  const [isLoading, setLoading] = React.useState(false);

  const getNewAddress = React.useCallback(async () => {
    if (!csrf) {
      return;
    }
    setLoading(true);
    try {
      const address = await request<string>("/api/wallet/new-address", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
      });
      if (!address) {
        throw new Error("No address in response");
      }
      swr.mutate(address, false);
    } catch (error) {
      alert("Failed to request a new address: " + error);
    } finally {
      setLoading(false);
    }
  }, [csrf, swr]);

  return React.useMemo(
    () => ({
      ...swr,
      getNewAddress,
      loadingAddress: isLoading || !swr.data,
    }),
    [swr, getNewAddress, isLoading]
  );
}
