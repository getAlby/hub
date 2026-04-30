import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import OnboardingChecklist from "src/screens/wallet/OnboardingChecklist";

import React from "react";
import { AlbyAccountWidget } from "src/components/home/widgets/AlbyAccountWidget";
import { AlbyExtensionWidget } from "src/components/home/widgets/AlbyExtensionWidget";
import { AlbyGoWidget } from "src/components/home/widgets/AlbyGoWidget";
import { AppOfTheDayWidget } from "src/components/home/widgets/AppOfTheDayWidget";
import { BlockHeightWidget } from "src/components/home/widgets/BlockHeightWidget";
import { ForwardsWidget } from "src/components/home/widgets/ForwardsWidget";
import { LatestUsedAppsWidget } from "src/components/home/widgets/LatestUsedAppsWidget";
import { LightningMessageboardWidget } from "src/components/home/widgets/LightningMessageboardWidget";
import { NewArrivalsWidget } from "src/components/home/widgets/NewArrivalsWidget";
import { NodeStatusWidget } from "src/components/home/widgets/NodeStatusWidget";
import { OnchainFeesWidget } from "src/components/home/widgets/OnchainFeesWidget";
import { SupportAlbyWidget } from "src/components/home/widgets/SupportAlbyWidget";
import { WhatsNewWidget } from "src/components/home/widgets/WhatsNewWidget";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { SearchInput } from "src/components/ui/search-input";

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

  if (!info || !balances) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader
        title={getGreeting(albyMe?.name)}
        pageTitle="Home"
        contentRight={<SearchInput placeholder="Search" />}
      />
      <div className="columns-1 lg:columns-2 gap-3 *:mb-3 *:break-inside-avoid">
        <OnboardingChecklist />
        <WhatsNewWidget />
        <LatestUsedAppsWidget />
        <NewArrivalsWidget />
        <AppOfTheDayWidget />
        <SupportAlbyWidget />
        <div className="flex flex-col gap-3">
          <AlbyAccountWidget />
          <AlbyGoWidget />
          <AlbyExtensionWidget />
        </div>
        <LightningMessageboardWidget />
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
              <div className="grid gap-3">
                <NodeStatusWidget />
                <BlockHeightWidget />
                <OnchainFeesWidget />
                <ForwardsWidget />
              </div>
            </CardContent>
          )}
        </Card>
      </div>
    </>
  );
}

export default Home;
