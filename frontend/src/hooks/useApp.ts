import useSWR, { SWRConfiguration } from "swr";

import { App, ConnectionIssue } from "src/types";
import { swrFetcher } from "src/utils/swr";

const pollConfiguration: SWRConfiguration = {
  refreshInterval: 3000,
};

export function useApp(id: number | undefined, poll = false) {
  return useSWR<App>(
    !!id && `/api/v2/apps/${id}`,
    swrFetcher,
    poll ? pollConfiguration : undefined
  );
}

export function useConnectionIssues(appId: number | undefined) {
  return useSWR<ConnectionIssue[]>(
    !!appId && `/api/v2/apps/${appId}/issues?limit=5`,
    swrFetcher,
    pollConfiguration
  );
}
