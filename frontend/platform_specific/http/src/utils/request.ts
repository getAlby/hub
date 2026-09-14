import { ErrorResponse } from "src/types";

export const request = async <T>(
  ...args: Parameters<typeof fetch>
): Promise<T | undefined> => {
  if (import.meta.env.BASE_URL !== "/") {
    // if running on a subpath, include the subpath in the request URL
    // BASE_URL is set via process.env.BASE_PATH, see https://vite.dev/guide/build#public-base-path
    args[0] = import.meta.env.BASE_URL + args[0];
  }

  // The session is an HttpOnly cookie, which fetch sends along with same-origin
  // requests by default. This leaves the Authorization header untouched so that
  // a reverse proxy in front of Alby Hub can use HTTP Basic Authentication.
  try {
    const fetchResponse = await fetch(...args);

    let body: T | undefined;
    if (fetchResponse.status !== 204) {
      try {
        body = await fetchResponse.json();
      } catch (error) {
        console.error(error);
      }
    }

    if (!fetchResponse.ok) {
      throw new Error(
        fetchResponse.status +
          " " +
          ((body as ErrorResponse)?.message || "Unknown error")
      );
    }
    return body;
  } catch (error) {
    console.error("Failed to fetch", error);
    throw error;
  }
};
