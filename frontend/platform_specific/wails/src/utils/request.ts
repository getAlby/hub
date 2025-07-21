import { WailsRequestRouter } from "wailsjs/go/wails/WailsApp";

export const request = async <T>(
  ...args: Parameters<typeof fetch>
): Promise<T | undefined> => {
  try {
    const res = await WailsRequestRouter(
      args[0].toString(),
      args[1]?.method || "GET",
      args[1]?.body?.toString() || ""
    );

    console.info("Wails request", ...args, res);
    if (res.error) {
      throw new Error(res.error);
    }

    return res.body;
  } catch (error) {
    console.error("Failed to fetch", error);
    throw error;
  }
};
