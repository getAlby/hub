/**
 * Adds the Hub JWT without replacing caller-provided headers, including
 * Authorization credentials used by an HTTPS reverse proxy.
 */
export const withAlbyAuthHeader = (
  headers: HeadersInit | undefined,
  token: string
): Headers => {
  const normalizedHeaders = new Headers(headers);
  normalizedHeaders.set("X-Alby-Auth", `Bearer ${token}`);
  return normalizedHeaders;
};
