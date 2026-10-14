// e.g. https://api.boltz.exchange -> boltz.exchange
export function getSwapProviderName(swapApiUrl: string) {
  try {
    return new URL(swapApiUrl).host.replace(/^api\./, "");
  } catch {
    return swapApiUrl;
  }
}
