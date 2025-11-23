import { OnchainTransaction } from "src/types";
import { swrFetcher } from "src/utils/swr";
import useSWR, { SWRConfiguration } from "swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 30000,
};

export function useOnchainTransactions() {
  return useSWR<OnchainTransaction[]>(
    "/api/node/transactions",
    swrFetcher,
    pollConfiguration
  );
}
