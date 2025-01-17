import dayjs from "dayjs";
import { ChevronRight } from "lucide-react";
import { Link } from "react-router-dom";
import AppAvatar from "src/components/AppAvatar";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useApps } from "src/hooks/useApps";

export function LatestUsedAppsWidget() {
  const { data: apps } = useApps();
  if (!apps) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Latest Used Apps</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-4">
        {apps
          .sort(
            (a, b) =>
              new Date(b.lastEventAt ?? 0).getTime() -
              new Date(a.lastEventAt ?? 0).getTime()
          )
          .slice(0, 3)
          .map((app) => (
            <Link key={app.id} to={`/apps/${app.appPubkey}`}>
              <div className="flex items-center justify-between w-full">
                <div className="flex items-center justify-center gap-4">
                  <AppAvatar app={app} className="w-14 h-14 rounded-lg" />
                  <p className="text-xl font-semibold">
                    {app.name === "getalby.com" ? "Alby Account" : app.name}
                  </p>
                </div>
                <div className="flex items-center justify-center gap-4">
                  <p className="text-sm text-muted-foreground">
                    {app.lastEventAt
                      ? dayjs(app.lastEventAt).fromNow()
                      : "never"}
                  </p>
                  <ChevronRight className="text-muted-foreground w-8 h-8" />
                </div>
              </div>
            </Link>
          ))}
      </CardContent>
    </Card>
  );
}
