import React from "react";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import { FeaturedAppWidget } from "./FeaturedAppWidget";

type Props = {
  app?: AppStoreApp;
};

export function AppOfTheDayWidget({ app: appProp }: Props) {
  // filter out apps which already have a widget
  const [fallbackApp] = React.useState(() => {
    const excludedAppIds = ["alby-go", "zapplanner"];
    const apps = appStoreApps.filter(
      (entry) => !excludedAppIds.includes(entry.id)
    );
    return apps[Math.floor(Math.random() * apps.length)];
  });
  const app = appProp ?? fallbackApp;

  return <FeaturedAppWidget title="App of the Day" app={app} />;
}
