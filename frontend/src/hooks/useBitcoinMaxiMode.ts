import React from "react";
import { localStorageKeys } from "src/constants";

export function useBitcoinMaxiMode() {
  const [bitcoinMaxiMode, setBitcoinMaxiModeState] = React.useState<boolean>(
    localStorage.getItem(localStorageKeys.bitcoinMaxiMode) === "true"
  );

  const setBitcoinMaxiMode = (enabled: boolean) => {
    localStorage.setItem(localStorageKeys.bitcoinMaxiMode, String(enabled));
    setBitcoinMaxiModeState(enabled);
  };

  return {
    bitcoinMaxiMode,
    setBitcoinMaxiMode,
  };
}
