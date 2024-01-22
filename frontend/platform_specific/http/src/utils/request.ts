import { ErrorResponse } from "src/types";

export const request = async <T>(
  ...args: Parameters<typeof fetch>
): Promise<T | undefined> => {
  try {
    const fetchResponse = await fetch(...args);

    let body: T | undefined;
    try {
      body = await fetchResponse.json();
    } catch (error) {
      console.error(error);
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
