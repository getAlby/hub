import { request } from "./request";

export const swrFetcher = async (...args: Parameters<typeof fetch>) => {
  return request(...args);
};
