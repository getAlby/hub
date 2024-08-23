import useSWRImmutable from "swr/immutable";

import React from "react";

import { useToast } from "src/components/ui/use-toast";
import { request } from "src/utils/request";
import { swrFetcher } from "src/utils/swr";

export function useOnchainAddress() {
  // Use useSWRImmutable to avoid address randomly changing after deposit (e.g. on page re-focus on the channel order page)
  const swr = useSWRImmutable<string>("/api/wallet/address", swrFetcher, {
    revalidateOnMount: true,
  });
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);

  const getNewAddress = React.useCallback(async () => {
    setLoading(true);
    try {
      const address = await request<string>("/api/wallet/new-address", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });
      if (!address) {
        throw new Error("No address in response");
      }
      swr.mutate(address, false);
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to request a new address",
        description: "" + error,
      });
    } finally {
      setLoading(false);
    }
  }, [swr, toast]);

  return React.useMemo(
    () => ({
      ...swr,
      getNewAddress,
      loadingAddress: isLoading || !swr.data,
    }),
    [swr, getNewAddress, isLoading]
  );
}
