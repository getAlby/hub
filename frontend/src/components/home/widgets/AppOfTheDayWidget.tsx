import { ExternalLinkIcon } from "lucide-react";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";

export function AppOfTheDayWidget() {
  function seededRandom(seed: number) {
    const x = Math.sin(seed++) * 10000;
    return x - Math.floor(x);
  }

  // filter out apps which already have a widget
  const excludedAppIds = ["alby-go", "zapplanner"];
  const apps = appStoreApps.filter((a) => !excludedAppIds.includes(a.id));

  const daysSinceEpoch = Math.floor(Date.now() / (1000 * 60 * 60 * 24));
  const todayIndex = Math.floor(seededRandom(daysSinceEpoch) * apps.length);
  const app = apps[todayIndex];

  return (
    <Card>
      <CardHeader>
        <CardTitle>App of the Day</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex gap-3 items-center">
          <img
            src={app.logo}
            alt="logo"
            className="inline rounded-lg w-12 h-12"
          />
          <div className="grow">
            <CardTitle>{app.title}</CardTitle>
            <CardDescription>{app.description}</CardDescription>
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex flex-row justify-end">
        <LinkButton
          to={app.internal ? `/internal-apps/${app.id}` : `/appstore/${app.id}`}
          variant="outline"
        >
          <ExternalLinkIcon />
          Open
        </LinkButton>
      </CardFooter>
    </Card>
  );
}
