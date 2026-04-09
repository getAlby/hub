import dayjs from "dayjs";
import { ChevronRightIcon } from "lucide-react";
import { Link } from "react-router-dom";
import AppAvatar from "src/components/AppAvatar";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useApps } from "src/hooks/useApps";
import { getAppDisplayName } from "src/lib/utils";

export function LatestUsedAppsWidget() {
  const { data: appsData } = useApps(
    3,
    undefined,
    undefined,
    "last_settled_transaction"
  );
  const apps = appsData?.apps;
  const usedApps = apps?.filter((x) => x.lastSettledTransactionAt);

  if (!usedApps?.length) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div>Recently Used Apps</div>
          <LinkButton to="/apps?tab=connected-apps" variant="secondary">
            See All
          </LinkButton>
        </CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-4">
        {usedApps.map((app) => (
          <Link key={app.id} to={`/apps/${app.id}`}>
            <div className="flex items-center w-full gap-4">
              <AppAvatar app={app} className="w-14 h-14 rounded-lg" />
              <p className="text-sm font-medium flex-1 truncate">
                {getAppDisplayName(app.name)}
              </p>
              <p className="text-xs text-muted-foreground">
                {app.lastSettledTransactionAt
                  ? dayjs(app.lastSettledTransactionAt).fromNow()
                  : "never"}
              </p>
              <ChevronRightIcon className="text-muted-foreground size-8" />
            </div>
          </Link>
        ))}
      </CardContent>
    </Card>
  );
}
