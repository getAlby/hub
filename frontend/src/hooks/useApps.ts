import useSWR from "swr";

import { LIST_APPS_LIMIT } from "src/constants";
import { ListAppsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

export function useApps(
  limit = LIST_APPS_LIMIT,
  page = 1,
  filters?: {
    name?: string;
    appStoreAppId?: string;
    unused?: boolean;
    subWallets?: boolean;
  },
  orderBy?: "last_used_at" | "created_at"
) {
  const offset = (page - 1) * limit;
  return useSWR<ListAppsResponse>(
    `/api/apps?limit=${limit}&offset=${offset}&filters=${JSON.stringify(filters || {})}&order_by=${orderBy || ""}`,
    swrFetcher
  );
}
