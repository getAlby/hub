import React from "react";
import { useSWRConfig } from "swr";
import { ListTransactionsResponse } from "src/types";

/**
 * Returns label keys the user has used previously, derived from any
 * /api/transactions responses currently in the SWR cache. Powers
 * autocomplete on the label key input. Empty array if the user has not
 * labeled any transaction yet.
 */
export function useLabelKeySuggestions(): string[] {
  const { cache } = useSWRConfig();

  return React.useMemo(() => {
    const keys = new Set<string>();
    for (const cacheKey of cache.keys()) {
      if (typeof cacheKey !== "string") {
        continue;
      }
      if (!cacheKey.startsWith("/api/transactions")) {
        continue;
      }
      const entry = cache.get(cacheKey);
      const data = entry?.data as ListTransactionsResponse | undefined;
      if (!data?.transactions) {
        continue;
      }
      for (const tx of data.transactions) {
        const userLabels = tx.metadata?.user_labels;
        if (!userLabels) {
          continue;
        }
        for (const key of Object.keys(userLabels)) {
          keys.add(key);
        }
      }
    }
    return Array.from(keys).sort();
  }, [cache]);
}
