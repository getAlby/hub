import useSWRImmutable from "swr/immutable";

import { AlbyAuthUrlResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useAlbyAuthUrl() {
  return useSWRImmutable<AlbyAuthUrlResponse>("/api/alby/auth-url", swrFetcher);
}
