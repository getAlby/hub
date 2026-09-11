import { getAuthToken } from "src/lib/auth";
import { ErrorResponse } from "src/types";

import { withAlbyAuthHeader } from "src/lib/authHeaders";

export const request = async <T>(
  ...args: Parameters<typeof fetch>
): Promise<T | undefined> => {
  if (import.meta.env.BASE_URL !== "/") {
    // if running on a subpath, include the subpath in the request URL
    // BASE_URL is set via process.env.BASE_PATH, see https://vite.dev/guide/build#public-base-path
    args[0] = import.meta.env.BASE_URL + args[0];
  }

  const token = getAuthToken();
  if (token) {
    if (!args[1]) {
      args[1] = {};
    }
    args[1].headers = withAlbyAuthHeader(args[1].headers, token);
    // Unlike Authorization, custom headers can be forwarded across origins
    // when fetch follows a redirect. API calls must never send the Hub token
    // to a redirect target.
    args[1].redirect = "error";
  }

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
