import { getAuthToken } from "src/lib/auth";
import { ErrorResponse } from "src/types";

const BASE_URL = import.meta.env.BASE_URL;
const PREFIXES = ["/api", "/images"];

function startsWithPrefix(path: string, prefixes: string[]): boolean {
  return prefixes.some((prefix) => path.startsWith(prefix));
}

export const request = async <T>(
  ...args: Parameters<typeof fetch>
): Promise<T | undefined> => {
  if (
    BASE_URL !== "/" &&
    typeof args[0] === "string" &&
    startsWithPrefix(args[0], PREFIXES)
  ) {
    args[0] = BASE_URL + args[0].slice(1);
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
