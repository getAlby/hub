import useSWR from "swr";

import { swrFetcher } from "src/swr";
import { User } from "src/types";

export function useUser() {
  return useSWR<User>("/api/user/me", swrFetcher);
}
