import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";

type BlogPost = {
  id: string;
  title: string;
  description: string;
  url: string;
  imageUrl?: string;
};

export function useAlbyBlog() {
  return useSWR<BlogPost>("/api/alby/blog/latest", swrFetcher, {
    dedupingInterval: 5 * 60 * 1000, // 5 minutes
  });
}
