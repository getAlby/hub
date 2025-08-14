import { GlobeIcon } from "lucide-react";
import { Link, useNavigate, useParams } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { ChromeIcon } from "src/components/icons/Chrome";
import { FirefoxIcon } from "src/components/icons/Firefox";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { ZapStoreIcon } from "src/components/icons/ZapStore";
import { suggestedApps } from "src/components/SuggestedAppData";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function AppStoreDetail() {
  const { appId } = useParams() as { appId: string };
  const app = suggestedApps.find((x) => x.id === appId);
  const navigate = useNavigate();

  if (!app) {
    navigate("/appstore");
    return;
  }

  // Redirect internal apps to their dedicated pages
  if (app.internal) {
    navigate(`/internal-apps/${appId}`);
    return;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title={
          <>
            <div className="flex flex-row items-center">
              <img src={app.logo} className="w-14 h-14 rounded-lg mr-4" />
              <div className="flex flex-col">
                <div>{app.title}</div>
                <div className="text-sm font-normal text-muted-foreground">
                  {app.description}
                </div>
              </div>
            </div>
          </>
        }
        description=""
        contentRight={
          <Link to={`/apps/new?app=${app.id}`}>
            <Button>
              <NostrWalletConnectIcon />
              Connect to {app.title}
            </Button>
          </Link>
        }
      />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="flex flex-col w-full gap-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">About the App</CardTitle>
            </CardHeader>
            {app.extendedDescription && (
              <CardContent className="flex flex-col gap-3">
                <p className="text-muted-foreground">
                  {app.extendedDescription}
                </p>
              </CardContent>
            )}
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">How to Connect</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              {app.guide || (
                <ul className="list-inside list-decimal">
                  <li>Install the app</li>
                  <li>
                    Click{" "}
                    <Link to={`/apps/new?app=${appId}`}>
                      <Button variant="link" className="px-0">
                        Connect to {app.title}
                      </Button>
                    </Link>
                  </li>
                  <li>
                    Open the Alby Go app on your mobile and scan the QR code
                  </li>
                </ul>
              )}
            </CardContent>
          </Card>
        </div>
        <div className="flex flex-col w-full gap-6">
          {(app.appleLink ||
            app.playLink ||
            app.zapStoreLink ||
            app.chromeLink ||
            app.firefoxLink) && (
            <Card>
              <CardHeader>
                <CardTitle className="text-2xl">Get This App</CardTitle>
              </CardHeader>
              <CardFooter className="flex flex-row gap-2">
                {app.playLink && (
                  <ExternalLink to={app.playLink}>
                    <Button variant="outline">
                      <PlayStoreIcon />
                      Play Store
                    </Button>
                  </ExternalLink>
                )}
                {app.appleLink && (
                  <ExternalLink to={app.appleLink}>
                    <Button variant="outline">
                      <AppleIcon />
                      App Store
                    </Button>
                  </ExternalLink>
                )}
                {app.zapStoreLink && (
                  <ExternalLink to={app.zapStoreLink}>
                    <Button variant="outline">
                      <ZapStoreIcon />
                      Zapstore
                    </Button>
                  </ExternalLink>
                )}
                {app.chromeLink && (
                  <ExternalLink to={app.chromeLink}>
                    <Button variant="outline">
                      <ChromeIcon />
                      Chrome Web Store
                    </Button>
                  </ExternalLink>
                )}
                {app.firefoxLink && (
                  <ExternalLink to={app.firefoxLink}>
                    <Button variant="outline">
                      <FirefoxIcon />
                      Firefox Add-Ons
                    </Button>
                  </ExternalLink>
                )}
              </CardFooter>
            </Card>
          )}
          {app.webLink && (
            <Card>
              <CardHeader>
                <CardTitle className="text-2xl">Links</CardTitle>
              </CardHeader>
              <CardFooter className="flex flex-row gap-2">
                {app.webLink && (
                  <ExternalLink to={app.webLink}>
                    <Button variant="outline">
                      <GlobeIcon />
                      Website
                    </Button>
                  </ExternalLink>
                )}
              </CardFooter>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
