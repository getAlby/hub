import { ErrorResponse } from "src/types";
import { appFetch } from "./fetch";

export const swrFetcher = async (...args: Parameters<typeof fetch>) => {
  const response = await appFetch(args);

  const json = await response.json();

  if (!response.ok) {
    const error = new Error(
      response.status +
        " " +
        ((json as ErrorResponse).message || "Unknown error")
    );
    throw error;
  }

  return json;
};
