import { getAuthToken } from "src/lib/auth";
import useSetupStore from "src/state/SetupStore";
import { ErrorResponse } from "src/types";

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
    args[1].headers = {
      ...args[1].headers,
      Authorization: `Bearer ${token}`,
    };
  }

  const setupCode = useSetupStore.getState().setupCode;
  if (setupCode) {
    if (!args[1]) {
      args[1] = {};
    }
    args[1].headers = {
      ...args[1].headers,
      "X-Setup-Code": setupCode,
    };
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
      if (
        setupCode &&
        fetchResponse.status === 401 &&
        (body as ErrorResponse)?.message === "Invalid setup code"
      ) {
        // ask the user to enter the setup code again
        useSetupStore.getState().setSetupCode("");
      }
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
