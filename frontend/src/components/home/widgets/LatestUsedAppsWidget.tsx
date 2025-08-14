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
import { ALBY_ACCOUNT_APP_NAME } from "src/constants";
import { useApps } from "src/hooks/useApps";

export function LatestUsedAppsWidget() {
  const { data: appsData } = useApps(3, undefined, undefined, "last_used_at");
  const apps = appsData?.apps;
  const usedApps = apps?.filter((x) => x.lastUsedAt);

  if (!usedApps?.length) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recently Used Apps</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-4">
        {usedApps
          .sort(
            (a, b) =>
              new Date(b.lastUsedAt ?? 0).getTime() -
              new Date(a.lastUsedAt ?? 0).getTime()
          )
          .map((app) => (
            <Link key={app.id} to={`/apps/${app.id}`}>
              <div className="flex items-center w-full gap-4">
                <AppAvatar app={app} className="w-14 h-14 rounded-lg" />
                <p className="text-sm font-medium flex-1 truncate">
                  {app.name === ALBY_ACCOUNT_APP_NAME
                    ? "Alby Account"
                    : app.name}
                </p>
                <p className="text-xs text-muted-foreground">
                  {app.lastUsedAt ? dayjs(app.lastUsedAt).fromNow() : "never"}
                </p>
                <ChevronRightIcon className="text-muted-foreground size-8" />
              </div>
            </Link>
          ))}
      </CardContent>
    </Card>
  );
}
