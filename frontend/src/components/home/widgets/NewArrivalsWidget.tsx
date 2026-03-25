import { ChevronRightIcon } from "lucide-react";
import { Link } from "react-router-dom";
import {
  AppStoreApp,
  appStoreApps,
} from "src/components/connections/SuggestedAppData";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

const latestAppStoreAppIds = [
  "goose",
  "claude",
  "alby-cli-skill",
  "bitrefill",
  "tictactoe",
  "wherostr",
] as const;

function getAppDestination(app: AppStoreApp) {
  return app.internal ? `/internal-apps/${app.id}` : `/appstore/${app.id}`;
}

export function NewArrivalsWidget() {
  const latestApps = latestAppStoreAppIds
    .map((id) => appStoreApps.find((app) => app.id === id))
    .filter((app): app is AppStoreApp => !!app)
    .slice(0, 3);

  if (!latestApps.length) {
    return null;
  }

  return (
    <Card className="overflow-hidden rounded-[14px] shadow-none">
      <CardHeader className="px-6 pb-0">
        <CardTitle className="text-base font-semibold">New Arrivals</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4 px-6 pt-0">
        {latestApps.map((app) => (
          <Link key={app.id} to={getAppDestination(app)} className="group">
            <div className="flex items-center gap-4">
              <img
                src={app.logo}
                alt={`${app.title} logo`}
                className="size-[60px] rounded-[9px] object-cover"
              />
              <div className="min-w-0 flex-1">
                <p className="truncate text-base font-semibold text-foreground">
                  {app.title}
                </p>
                <p className="line-clamp-2 text-sm leading-5 text-muted-foreground">
                  {app.description}
                </p>
              </div>
              <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
            </div>
          </Link>
        ))}
      </CardContent>
    </Card>
  );
}
