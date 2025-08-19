import { GlobeIcon } from "lucide-react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { AppDetailConnectedApps } from "src/components/connections/AppDetailConnectedApps";
import { AppDetailHeader } from "src/components/connections/AppDetailHeader";
import { AppTransactionList } from "src/components/connections/AppTransactionList";
import { AppUsage } from "src/components/connections/AppUsage";
import {
  SuggestedApp,
  suggestedApps,
} from "src/components/connections/SuggestedAppData";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { ChromeIcon } from "src/components/icons/Chrome";
import { FirefoxIcon } from "src/components/icons/Firefox";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { ZapStoreIcon } from "src/components/icons/ZapStore";
import Permissions from "src/components/Permissions";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useAppsForAppStoreApp } from "src/hooks/useApps";
import { useCapabilities } from "src/hooks/useCapabilities";

export function AppStoreDetail() {
  const { appStoreId } = useParams() as { appStoreId: string };
  const appStoreApp = suggestedApps.find((x) => x.id === appStoreId);
  const navigate = useNavigate();

  if (!appStoreApp) {
    navigate("/apps?tab=app-store");
    return null;
  }

  // Redirect internal apps to their dedicated pages
  if (appStoreApp.internal) {
    navigate(`/internal-apps/${appStoreId}`);
    return null;
  }

  return (
    <AppStoreDetailInternal appStoreId={appStoreId} appStoreApp={appStoreApp} />
  );
}

function AppStoreDetailInternal({
  appStoreApp,
  appStoreId,
}: {
  appStoreApp: SuggestedApp;
  appStoreId: string;
}) {
  const connectedApps = useAppsForAppStoreApp(appStoreApp);
  const { data: capabilities } = useCapabilities();

  return (
    <div className="grid gap-5">
      <AppDetailHeader appStoreApp={appStoreApp} />
      {connectedApps.length > 1 && (
        <AppDetailConnectedApps appStoreApp={appStoreApp} />
      )}
      {connectedApps.length === 1 && (
        <>
          <AppUsage app={connectedApps[0]} />
          {capabilities && (
            <Card>
              <CardHeader>
                <CardTitle>Permissions</CardTitle>
              </CardHeader>
              <CardContent>
                <Permissions
                  capabilities={capabilities}
                  permissions={{
                    scopes: connectedApps[0].scopes,
                    maxAmount: connectedApps[0].maxAmount,
                    budgetRenewal: connectedApps[0].budgetRenewal,
                    expiresAt: connectedApps[0].expiresAt
                      ? new Date(connectedApps[0].expiresAt)
                      : undefined,
                    isolated: connectedApps[0].isolated,
                  }}
                  readOnly
                  isNewConnection={false}
                  budgetUsage={connectedApps[0].budgetUsage}
                />
              </CardContent>
            </Card>
          )}
          <AppTransactionList appId={connectedApps[0].id} />
        </>
      )}

      {connectedApps.length === 0 && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="flex flex-col w-full gap-6">
              <Card>
                <CardHeader>
                  <CardTitle className="text-2xl">About the App</CardTitle>
                </CardHeader>
                {appStoreApp.extendedDescription && (
                  <CardContent className="flex flex-col gap-3">
                    <p className="text-muted-foreground">
                      {appStoreApp.extendedDescription}
                    </p>
                  </CardContent>
                )}
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="text-2xl">How to Connect</CardTitle>
                </CardHeader>
                <CardContent className="flex flex-col gap-3">
                  {appStoreApp.guide || (
                    <ul className="list-inside list-decimal">
                      <li>Install the app</li>
                      <li>
                        Click{" "}
                        <Link to={`/apps/new?app=${appStoreId}`}>
                          <Button variant="link" className="px-0">
                            Connect to {appStoreApp.title}
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
              {(appStoreApp.appleLink ||
                appStoreApp.playLink ||
                appStoreApp.zapStoreLink ||
                appStoreApp.chromeLink ||
                appStoreApp.firefoxLink) && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-2xl">Get This App</CardTitle>
                  </CardHeader>
                  <CardFooter className="flex flex-row gap-2">
                    {appStoreApp.playLink && (
                      <ExternalLink to={appStoreApp.playLink}>
                        <Button variant="outline">
                          <PlayStoreIcon />
                          Play Store
                        </Button>
                      </ExternalLink>
                    )}
                    {appStoreApp.appleLink && (
                      <ExternalLink to={appStoreApp.appleLink}>
                        <Button variant="outline">
                          <AppleIcon />
                          App Store
                        </Button>
                      </ExternalLink>
                    )}
                    {appStoreApp.zapStoreLink && (
                      <ExternalLink to={appStoreApp.zapStoreLink}>
                        <Button variant="outline">
                          <ZapStoreIcon />
                          Zapstore
                        </Button>
                      </ExternalLink>
                    )}
                    {appStoreApp.chromeLink && (
                      <ExternalLink to={appStoreApp.chromeLink}>
                        <Button variant="outline">
                          <ChromeIcon />
                          Chrome Web Store
                        </Button>
                      </ExternalLink>
                    )}
                    {appStoreApp.firefoxLink && (
                      <ExternalLink to={appStoreApp.firefoxLink}>
                        <Button variant="outline">
                          <FirefoxIcon />
                          Firefox Add-Ons
                        </Button>
                      </ExternalLink>
                    )}
                  </CardFooter>
                </Card>
              )}
              {appStoreApp.webLink && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-2xl">Links</CardTitle>
                  </CardHeader>
                  <CardFooter className="flex flex-row gap-2">
                    {appStoreApp.webLink && (
                      <ExternalLink to={appStoreApp.webLink}>
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
        </>
      )}

      {connectedApps.length !== 0 && <></>}
    </div>
  );
}
