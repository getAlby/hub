import { request } from "./request";

export const swrFetcher = async (...args: Parameters<typeof fetch>) => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return request(...args) as any;
};
