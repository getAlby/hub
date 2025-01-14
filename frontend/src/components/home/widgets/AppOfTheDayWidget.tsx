import { ExternalLinkIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { suggestedApps } from "src/components/SuggestedAppData";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function AppOfTheDayWidget() {
  function seededRandom(seed: number) {
    seed = (seed * 48271) % 2147483647; // LCG formula
    return (seed - 1) / 2147483646; // Normalize to [0, 1)
  }

  const daysSinceEpoch = Math.floor(Date.now() / (1000 * 60 * 60 * 24));
  const todayIndex = Math.floor(
    seededRandom(daysSinceEpoch) * suggestedApps.length
  );
  const app = suggestedApps[todayIndex];

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
          <div className="flex-grow">
            <CardTitle>{app.title}</CardTitle>
            <CardDescription>{app.description}</CardDescription>
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex flex-row justify-end">
        <Link
          to={app.internal ? `/internal-apps/${app.id}` : `/appstore/${app.id}`}
        >
          <Button variant="outline">
            <ExternalLinkIcon className="w-4 h-4 mr-2" />
            Open
          </Button>
        </Link>
      </CardFooter>
    </Card>
  );
}
