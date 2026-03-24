import React from "react";
import AppHeader from "src/components/AppHeader";
import {
  AppStoreApp,
  appStoreApps,
  latestAppStoreAppIds,
} from "src/components/connections/SuggestedAppData";
import { AlbyBlogWidget } from "src/components/home/widgets/AlbyBlogWidget";
import { AppOfTheDayWidget } from "src/components/home/widgets/AppOfTheDayWidget";
import { HomeTopChartsRow } from "src/components/home/widgets/HomeTopChartsRow";
import { LatestUsedAppsWidget } from "src/components/home/widgets/LatestUsedAppsWidget";
import { NewArrivalWidget } from "src/components/home/widgets/NewArrivalWidget";
import { StoriesWidget } from "src/components/home/widgets/StoriesWidget";
import { SupportAlbyWidget } from "src/components/home/widgets/SupportAlbyWidget";
import { SearchInput } from "src/components/ui/search-input";
import OnboardingChecklist from "src/screens/wallet/OnboardingChecklist";

const excludedAppOfTheDayIds = ["alby-go", "zapplanner"];

function pickRandomApp(apps: AppStoreApp[]) {
  if (!apps.length) {
    return undefined;
  }

  return apps[Math.floor(Math.random() * apps.length)];
}

function Home() {
  const [latestArrivalApp] = React.useState(() =>
    pickRandomApp(
      latestAppStoreAppIds
        .map((id) => appStoreApps.find((app) => app.id === id))
        .filter((app): app is AppStoreApp => !!app)
    )
  );
  const [appOfTheDay] = React.useState(() =>
    pickRandomApp(
      appStoreApps.filter(
        (app) =>
          !excludedAppOfTheDayIds.includes(app.id) &&
          app.id !== latestArrivalApp?.id
      )
    )
  );

  return (
    <>
      <AppHeader
        title="Home"
        contentRight={<SearchInput placeholder="Search" />}
      />
      <HomeTopChartsRow />
      <div className="grid grid-cols-1 items-start gap-3 xl:grid-cols-2">
        <div className="grid min-w-0 gap-3">
          <OnboardingChecklist />
          <LatestUsedAppsWidget />
          <NewArrivalWidget app={latestArrivalApp} />
          <AppOfTheDayWidget app={appOfTheDay} />
        </div>
        <div className="grid min-w-0 gap-3">
          <StoriesWidget />
          <SupportAlbyWidget />
          <AlbyBlogWidget />
        </div>
      </div>
    </>
  );
}

export default Home;
