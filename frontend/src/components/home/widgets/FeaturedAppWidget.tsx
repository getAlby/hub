import { ChevronRightIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

type Props = {
  title: string;
  app?: AppStoreApp;
};

function getAppDestination(app: AppStoreApp) {
  return app.internal ? `/internal-apps/${app.id}` : `/appstore/${app.id}`;
}

export function FeaturedAppWidget({ title, app }: Props) {
  if (!app) {
    return null;
  }

  return (
    <Card className="overflow-hidden rounded-[14px] shadow-none">
      <CardHeader className="px-6 pb-0">
        <CardTitle className="text-base font-semibold">{title}</CardTitle>
      </CardHeader>
      <CardContent className="px-6 pt-0">
        <Link
          to={getAppDestination(app)}
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
            <p className="line-clamp-2 text-sm leading-5 text-muted-foreground">
              {app.description}
            </p>
          </div>
          <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
        </Link>
      </CardContent>
    </Card>
  );
}
