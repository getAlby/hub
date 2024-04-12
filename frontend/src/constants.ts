export const localStorageKeys = {
  returnTo: "returnTo",
  onchainAddress: "onchainAddress",
  onboardingSkipped: "onboardingSkipped",
};

export const MIN_0CONF_BALANCE = 30000; // 30,000 for Alby. 20000 works for Olympus and Voltage
export const ALBY_SERVICE_FEE = 8 / 1000;
export const ALBY_MIN_BALANCE = Math.ceil(
  MIN_0CONF_BALANCE * (1 + ALBY_SERVICE_FEE)
);

export const ONCHAIN_DUST_SATS = 1000;
