import useSWR from "swr";

import { swrFetcher } from "@swr";
import { User } from "@types";

export function useUser() {
  return useSWR<User>("/api/user/me", swrFetcher);
}
