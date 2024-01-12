import { ErrorResponse } from "src/types";

export const swrFetcher = async (...args: Parameters<typeof fetch>) => {
  const response = await fetch(...args);

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
