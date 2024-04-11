export const localStorageKeys = {
  returnTo: "returnTo",
  onchainAddress: "onchainAddress",
};

export const MIN_0CONF_BALANCE = 20000;
export const ALBY_MIN_0CONF_BALANCE = 30000;
export const ALBY_SERVICE_FEE = 8 / 1000;
export const ALBY_FEE_RESERVE = 0.01; /* TODO: remove when no fee reserve is needed */
export const MIN_ALBY_BALANCE =
  ALBY_MIN_0CONF_BALANCE +
  1000; /* TODO: remove when no fee reserve is needed */

export const ONCHAIN_DUST_SATS = 1000;
