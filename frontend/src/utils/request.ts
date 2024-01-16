import toast from "src/components/Toast";
import { ErrorResponse } from "src/types";

import { WailsRequestRouter } from "wailsjs/go/main/WailsApp";

export const request = async <T>(
  ...args: Parameters<typeof fetch>
): Promise<T | undefined> => {
  try {
    switch (import.meta.env.VITE_APP_TYPE) {
      case "WAILS": {
        const res = await WailsRequestRouter(args[0].toString());
        // TODO: wrap response and do error handling e.g.
        // if (!res.ok) { throw new Error((json as ErrorResponse).message || "Unknown error")}

        return res;
      }
      case "HTTP": {
        const fetchResponse = await fetch(...args);

        let json: T | undefined;
        try {
          json = await fetchResponse.json();
        } catch (error) {
          console.error(error);
        }

        if (!fetchResponse.ok) {
          throw new Error(
            fetchResponse.status +
              " " +
              ((json as ErrorResponse)?.message || "Unknown error")
          );
        }
        return json;
      }
      default:
        throw new Error(
          "Unsupported app type: " + import.meta.env.VITE_APP_TYPE
        );
    }
  } catch (error) {
    console.error("Failed to fetch", error);
    throw error;
  }
};

export function handleFetchError(message: string, error: unknown) {
  console.error(message, error);
  toast.error(message + ": " + error);
}
