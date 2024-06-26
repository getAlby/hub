import { Globe } from "lucide-react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardTitle,
} from "src/components/ui/card";
import { SuggestedApp, suggestedApps } from "./SuggestedAppData";

function SuggestedAppCard({
  id,
  title,
  description,
  logo,
  webLink,
  appleLink,
  playLink,
}: SuggestedApp) {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex gap-3 items-center">
          <img src={logo} alt="logo" className="inline rounded-lg w-12 h-12" />
          <div className="flex-grow">
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex flex-row justify-between">
        <div className="flex flex-row gap-4">
          <ExternalLink to={webLink}>
            <Button variant="outline" size="icon">
              <Globe className="w-4 h-4" />
            </Button>
          </ExternalLink>
          {appleLink && (
            <ExternalLink to={appleLink}>
              <Button variant="outline" size="icon">
                <AppleIcon className="w-4 h-4" />
              </Button>
            </ExternalLink>
          )}
          {playLink && (
            <ExternalLink to={playLink}>
              <Button variant="outline" size="icon">
                <PlayStoreIcon className="w-4 h-4" />
              </Button>
            </ExternalLink>
          )}
        </div>
        <Link to={`/apps/new?app=${id}`}>
          <Button variant="outline">
            <NostrWalletConnectIcon className="w-4 h-4 mr-2" />
            Connect
          </Button>
        </Link>
      </CardFooter>
    </Card>
  );
}

export default function SuggestedApps() {
  return (
    <>
      <div className="grid sm:grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {suggestedApps.map((app) => (
          <SuggestedAppCard key={app.id} {...app} />
        ))}
      </div>
    </>
  );
}
