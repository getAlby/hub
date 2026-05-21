import { ChevronRightIcon } from "lucide-react";
import { Link } from "react-router";
import {
  appStoreApps,
  getAppStoreUrl,
} from "src/components/connections/SuggestedAppData";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function NewArrivalsWidget() {
  const latestApps = appStoreApps
    .filter((app) => !!app.addedDate)
    .sort((a, b) => b.addedDate!.localeCompare(a.addedDate!))
    .slice(0, 3);

  if (!latestApps.length) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>New Arrivals</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4">
        {latestApps.map((app) => (
          <Link key={app.id} to={getAppStoreUrl(app)} className="group">
            <div className="flex items-center gap-3">
              <img
                src={app.logo}
                alt={`${app.title} logo`}
                className="w-12 h-12 rounded-lg object-cover"
              />
              <div className="min-w-0 flex-1">
                <CardTitle>{app.title}</CardTitle>
                <p className="text-sm text-muted-foreground line-clamp-1">
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
