import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import { FeaturedAppWidget } from "./FeaturedAppWidget";

type Props = {
  app?: AppStoreApp;
};

export function NewArrivalWidget({ app }: Props) {
  return <FeaturedAppWidget title="New Arrival" app={app} />;
}
