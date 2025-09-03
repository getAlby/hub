import { ExternalLinkIcon } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AlbyHead } from "src/components/images/AlbyHead";
import Loading from "src/components/Loading";
import StoriesSection from "src/components/StoriesSection";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import OnboardingChecklist from "src/screens/wallet/OnboardingChecklist";

import React from "react";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import zapplanner from "src/assets/suggested-apps/zapplanner.png";
import { AppOfTheDayWidget } from "src/components/home/widgets/AppOfTheDayWidget";
import { BlockHeightWidget } from "src/components/home/widgets/BlockHeightWidget";
import { ForwardsWidget } from "src/components/home/widgets/ForwardsWidget";
import { LatestUsedAppsWidget } from "src/components/home/widgets/LatestUsedAppsWidget";
import { LightningMessageboardWidget } from "src/components/home/widgets/LightningMessageboardWidget";
import { NodeStatusWidget } from "src/components/home/widgets/NodeStatusWidget";
import { OnchainFeesWidget } from "src/components/home/widgets/OnchainFeesWidget";
import { SupportAlbyWidget } from "src/components/home/widgets/SupportAlbyWidget";
import { WhatsNewWidget } from "src/components/home/widgets/WhatsNewWidget";

function getGreeting(name: string | undefined) {
  const hours = new Date().getHours();
  let greeting;

  if (hours < 11) {
    greeting = "Good Morning";
  } else if (hours < 16) {
    greeting = "Good Afternoon";
  } else {
    greeting = "Good Evening";
  }

  return `${greeting}${name ? `, ${name}` : ""}!`;
}

function Home() {
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
  const { data: albyMe } = useAlbyMe();
  const [isNerd, setNerd] = React.useState(false);
  /* eslint-disable  @typescript-eslint/no-explicit-any */
  const extensionInstalled = (window as any).alby !== undefined;

  if (!info || !balances) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader title={getGreeting(albyMe?.name)} />
      <StoriesSection />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 items-start justify-start">
        {/* LEFT */}
        <div className="grid gap-5">
          <OnboardingChecklist />
          <WhatsNewWidget />
          <SupportAlbyWidget />
          {info.albyAccountConnected && (
            <ExternalLink to="https://www.getalby.com/dashboard">
              <Card>
                <CardHeader>
                  <div className="flex flex-row items-center">
                    <div className="shrink-0">
                      <AlbyHead className="w-12 h-12 rounded-xl p-1 border" />
                    </div>
                    <div>
                      <CardTitle>
                        <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                          Alby Account
                        </div>
                      </CardTitle>
                      <CardDescription className="ml-4">
                        Get an Alby Account with a web wallet interface,
                        lightning address and other features.
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="text-right">
                  <Button variant="outline">
                    Open Alby Account
                    <ExternalLinkIcon className="size-4 ml-2" />
                  </Button>
                </CardContent>
              </Card>
            </ExternalLink>
          )}

          <Link to="/internal-apps/alby-go">
            <Card>
              <CardHeader>
                <div className="flex flex-row items-center">
                  <div className="shrink-0">
                    <img src={albyGo} className="w-12 h-12 rounded-xl border" />
                  </div>
                  <div>
                    <CardTitle>
                      <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                        Alby Go
                      </div>
                    </CardTitle>
                    <CardDescription className="ml-4">
                      The easiest Bitcoin mobile app that works great with Alby
                      Hub.
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="text-right">
                <Button variant="outline">Open</Button>
              </CardContent>
            </Card>
          </Link>
          {!extensionInstalled && (
            <ExternalLink to="https://getalby.com/products/browser-extension">
              <Card>
                <CardHeader>
                  <div className="flex flex-row items-center">
                    <div className="shrink-0">
                      <AlbyHead className="w-12 h-12 rounded-xl p-1 border bg-[#FFDF6F]" />
                    </div>
                    <div>
                      <CardTitle>
                        <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                          Alby Browser Extension
                        </div>
                      </CardTitle>
                      <CardDescription className="ml-4">
                        Seamless bitcoin payments in your favorite internet
                        browser.
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="text-right">
                  <Button variant="outline">
                    Install Alby Extension
                    <ExternalLinkIcon className="size-4 ml-2" />
                  </Button>
                </CardContent>
              </Card>
            </ExternalLink>
          )}
        </div>

        {/* RIGHT */}
        <div className="grid gap-5">
          <LatestUsedAppsWidget />
          <LightningMessageboardWidget />
          <AppOfTheDayWidget />

          <Link to="/internal-apps/zapplanner">
            <Card>
              <CardHeader>
                <div className="flex flex-row items-center">
                  <div className="shrink-0">
                    <img
                      src={zapplanner}
                      className="w-12 h-12 rounded-xl border"
                    />
                  </div>
                  <div>
                    <CardTitle>
                      <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4 flex gap-2">
                        ZapPlanner <Badge>NEW</Badge>
                      </div>
                    </CardTitle>
                    <CardDescription className="ml-4">
                      Schedule automatic recurring lightning payments.
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="text-right">
                <Button variant="outline">Open</Button>
              </CardContent>
            </Card>
          </Link>

          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>Stats for nerds</CardTitle>
                <Button variant="secondary" onClick={() => setNerd(!isNerd)}>
                  {isNerd ? "Hide" : "Show"}
                </Button>
              </div>
            </CardHeader>
            {isNerd && (
              <CardContent>
                <div className="grid gap-5">
                  <NodeStatusWidget />
                  <BlockHeightWidget />
                  <OnchainFeesWidget />
                  <ForwardsWidget />
                </div>
              </CardContent>
            )}
          </Card>
        </div>
      </div>
    </>
  );
}

export default Home;
