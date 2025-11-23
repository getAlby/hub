import useSWR from "swr";

import { GetForwardsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useForwards() {
  return useSWR<GetForwardsResponse>("/api/forwards", swrFetcher);
}
