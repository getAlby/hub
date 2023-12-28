export const swrFetcher = (...args: Parameters<typeof fetch>) =>
  fetch(...args).then((res) => res.json());
