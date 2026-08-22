// Parses a return_to URL and only returns it if it is a
// http or https URL. Relative URLs are resolved against the
// current origin.
export function safeReturnToUrl(returnTo: string | null): string | undefined {
  if (!returnTo) {
    return undefined;
  }
  try {
    const url = new URL(returnTo, window.location.origin);
    if (url.protocol === "http:" || url.protocol === "https:") {
      return url.toString();
    }
  } catch {
    // ignore invalid URLs
  }
  return undefined;
}
