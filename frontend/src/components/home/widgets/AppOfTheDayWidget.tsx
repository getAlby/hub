import { ChevronRightIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function AppOfTheDayWidget() {
  function seededRandom(seed: number) {
    const x = Math.sin(seed) * 10000;
    return x - Math.floor(x);
  }

  // filter out apps which already have a widget
  const excludedAppIds = ["alby-go", "zapplanner"];
  const apps = appStoreApps.filter((a) => !excludedAppIds.includes(a.id));

  // eslint-disable-next-line react-hooks/purity
  const daysSinceEpoch = Math.floor(Date.now() / (1000 * 60 * 60 * 24));
  const todayIndex = Math.floor(seededRandom(daysSinceEpoch) * apps.length);
  const app = apps[todayIndex];

  return (
    <Card className="overflow-hidden rounded-[14px] shadow-none">
      <CardHeader className="px-6 pb-0">
        <CardTitle className="text-base font-semibold">
          App of the Day
        </CardTitle>
      </CardHeader>
      <CardContent className="px-6 pt-0">
        <Link
          to={app.internal ? `/internal-apps/${app.id}` : `/appstore/${app.id}`}
          className="group flex items-center gap-4 rounded-md"
        >
          <img
            src={app.logo}
            alt={`${app.title} logo`}
            className="size-[60px] rounded-[9px] object-cover"
          />
          <div className="min-w-0 flex-1">
            <p className="truncate text-base font-semibold text-foreground">
              {app.title}
            </p>
            <CardDescription className="line-clamp-2 text-sm leading-5 text-muted-foreground">
              {app.description}
            </CardDescription>
          </div>
          <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
        </Link>
      </CardContent>
    </Card>
  );
}
