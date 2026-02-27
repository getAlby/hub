import React from "react";
import { useInfo } from "src/hooks/useInfo";
export function useBitcoinMaxiMode() {
  const { data: info } = useInfo();
  return React.useMemo(
    () => ({
      bitcoinMaxiMode: info?.bitcoinMaxiMode ?? false,
    }),
    [info?.bitcoinMaxiMode]
  );
}
