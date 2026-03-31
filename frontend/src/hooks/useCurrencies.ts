import React from "react";
import useSWR from "swr";

import { RATES_API_URL } from "src/constants";
import { handleRequestError } from "src/utils/handleRequestError";

const albyRatesFetcher = (url: string) =>
  fetch(url).then((res) => {
    if (!res.ok) {
      throw new Error(`Failed to fetch currencies: ${res.status}`);
    }
    return res.json() as Promise<
      Record<string, { name: string; priority: number }>
    >;
  });

export function useCurrencies(includeSats = false) {
  const { data: ratesData, isLoading } = useSWR<
    Record<string, { name: string; priority: number }>
  >(RATES_API_URL, albyRatesFetcher, {
    onError: (error) => handleRequestError("Failed to fetch currencies", error),
  });

  const currencies = React.useMemo(() => {
    if (!ratesData) {
      return [];
    }

    if (includeSats) {
      return [
        ["SATS", "sats"],
        ...Object.entries(ratesData)
          .filter(([code]) => code !== "BTC")
          .sort((a, b) => {
            const priorityDiff = a[1].priority - b[1].priority;
            return priorityDiff !== 0 ? priorityDiff : a[0].localeCompare(b[0]);
          })
          .map(([code, details]): [string, string] => [
            code.toUpperCase(),
            details.name,
          ]),
      ];
    }

    return Object.entries(ratesData)
      .filter(([code]) => code !== "BTC")
      .map(([code, details]): [string, string] => [
        code.toUpperCase(),
        details.name,
      ])
      .sort((a, b) => a[1].localeCompare(b[1]));
  }, [ratesData, includeSats]);

  return {
    currencies,
    isLoading,
  };
}
