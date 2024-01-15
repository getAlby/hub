import toast from "src/components/Toast";

// TODO: this should be passed as an environment variable
const FETCH_METHOD = "wails";

import { WailsRequestRouter } from "wailsjs/go/main/WailsApp";

export const appFetch = async (args: Parameters<typeof fetch>) => {
  try {
    if (FETCH_METHOD === "wails") {
      while (!("go" in window)) {
        console.log("go not in window");
        await new Promise((resolve) => setTimeout(resolve, 500));
      }
      const res = await WailsRequestRouter(args[0].toString());

      return {
        ok: true,
        status: 200,
        json: () => Promise.resolve(res),
      };
    }

    return fetch(...args);
  } catch (error) {
    console.error("Failed to fetch", error);
    throw error;
  }
};

export async function validateFetchResponse(response: Response) {
  if (!response.ok) {
    let reason = "unknown";
    try {
      reason = await response.text();
    } catch (error) {
      console.error("Failed to read response text", error);
    }
    throw new Error("Unexpected response: " + response.status + " " + reason);
  }
}

export function handleFetchError(message: string, error: unknown) {
  console.error(message, error);
  toast.error(message + ": " + error);
}
