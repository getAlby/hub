import { Globe } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { suggestedApps } from "src/components/SuggestedAppData";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

const ALBY_GO_APP_ID = "alby-go";

export function AlbyGo() {
  const app = suggestedApps.find((x) => x.id === ALBY_GO_APP_ID);

  return (
    <div className="grid gap-5">
      <AppHeader
        title={
          <>
            <div className="flex flex-row items-center">
              <img src={app?.logo} className="w-14 h-14 rounded-lg mr-4" />
              <div className="flex flex-col">
                <div>{app?.title}</div>
                <div className="text-sm font-normal text-muted-foreground">
                  {app?.description}
                </div>
              </div>
            </div>
          </>
        }
        description=""
        contentRight={
          <Link to={`/apps/new?app=${app?.id}`}>
            <Button variant="outline">
              <NostrWalletConnectIcon className="w-4 h-4 mr-2" />
              Connect to {app?.title}
            </Button>
          </Link>
        }
      />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Get This App</CardTitle>
          </CardHeader>
          <CardFooter className="flex flex-row gap-2">
            {app?.playLink && (
              <ExternalLink to={app?.playLink}>
                <Button variant="outline">
                  <PlayStoreIcon className="w-4 h-4 mr-2" />
                  Play Store
                </Button>
              </ExternalLink>
            )}

            {app?.appleLink && (
              <ExternalLink to={app.appleLink}>
                <Button variant="outline">
                  <AppleIcon className="w-4 h-4 mr-2" />
                  App Store
                </Button>
              </ExternalLink>
            )}
            <ExternalLink to="https://zap.store/download">
              <Button variant="outline">zap.store</Button>
            </ExternalLink>
          </CardFooter>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Links</CardTitle>
          </CardHeader>
          <CardFooter className="flex flex-row gap-2">
            {app?.webLink && (
              <ExternalLink to={app.webLink}>
                <Button variant="outline">
                  <Globe className="w-4 h-4 mr-2" />
                  Website
                </Button>
              </ExternalLink>
            )}
          </CardFooter>
        </Card>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>How to Connect</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="list-inside list-decimal">
            <li>Download the app from the app store</li>
            <li>
              Create a new app connection for Alby Go by clicking on{" "}
              <Link to="/apps/new?app=alby-go">
                <Button variant="link">Connect to Alby Go</Button>
              </Link>
            </li>
            <li>Open the Alby Go app on your mobile and scan the QR code</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
