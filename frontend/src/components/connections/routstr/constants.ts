export const ROUTSTR_APP_STORE_ID = "routstr";

export const ROUTSTR_UNIVERSAL_KEY_COPY =
  "This key works with all models available. Choose the model in your app when you call the API.";

export const ROUTSTR_AUTO_ROUTING_LABEL =
  "Auto-routing · cheapest provider at request time";

export const ROUTSTR_LOCAL_ENDPOINT = "http://localhost:8008/v1";

/** Path prefix on Hub for OpenAI-compatible proxy (no trailing slash). */
export const ROUTSTR_HUB_PROXY_PATH = "/routstr/v1";

export function getRoutstrHubEndpoint(): string {
  if (typeof window === "undefined") {
    return ROUTSTR_HUB_PROXY_PATH;
  }
  return `${window.location.origin}${ROUTSTR_HUB_PROXY_PATH}`;
}

export function isHubOpenedLocally(): boolean {
  if (typeof window === "undefined") {
    return true;
  }
  const host = window.location.hostname;
  return host === "localhost" || host === "127.0.0.1";
}
