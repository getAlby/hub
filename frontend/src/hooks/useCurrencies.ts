import React from "react";
import useSWR from "swr";

import { Currency } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useCurrencies(includeSats = false) {
  const { data: ratesData, isLoading } = useSWR<Currency[]>(
    "/api/alby/currencies",
    swrFetcher
  );

  const currencies = React.useMemo(() => {
    if (!ratesData) {
      return [];
    }

    if (includeSats) {
      return [
        ["SATS", "sats"],
        ...ratesData
          .filter(({ iso_code }) => iso_code !== "BTC")
          .sort((a, b) => {
            const priorityDiff = a.priority - b.priority;
            return priorityDiff !== 0
              ? priorityDiff
              : a.iso_code.localeCompare(b.iso_code);
          })
          .map(({ iso_code, name }): [string, string] => [iso_code, name]),
      ];
    }

    return ratesData
      .filter(({ iso_code }) => iso_code !== "BTC")
      .map(({ iso_code, name }): [string, string] => [iso_code, name])
      .sort((a, b) => a[1].localeCompare(b[1]));
  }, [ratesData, includeSats]);

  return {
    currencies,
    isLoading,
  };
}
