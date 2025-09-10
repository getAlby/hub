import { GlobeIcon } from "lucide-react";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { ChromeIcon } from "src/components/icons/Chrome";
import { FirefoxIcon } from "src/components/icons/Firefox";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { ZapStoreIcon } from "src/components/icons/ZapStore";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function AppLinksCard({ appStoreApp }: { appStoreApp: AppStoreApp }) {
  if (
    !appStoreApp.appleLink &&
    !appStoreApp.playLink &&
    !appStoreApp.zapStoreLink &&
    !appStoreApp.chromeLink &&
    !appStoreApp.firefoxLink &&
    !appStoreApp.webLink
  ) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-2xl">Links</CardTitle>
      </CardHeader>
      <CardFooter className="flex flex-row flex-wrap gap-2">
        {appStoreApp.webLink && (
          <ExternalLink to={appStoreApp.webLink}>
            <Button variant="outline">
              <GlobeIcon />
              Website
            </Button>
          </ExternalLink>
        )}
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
  );
}
