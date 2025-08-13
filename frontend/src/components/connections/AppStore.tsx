import { CirclePlusIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import ResponsiveButton from "src/components/ResponsiveButton";
import SuggestedApps from "src/components/connections/SuggestedApps";

function AppStore() {
  return (
    <>
      <AppHeader
        title="App Store"
        titleClassName="text-xl lg:text-2xl"
        contentRight={
          <>
            <ExternalLink to="https://github.com/getAlby/hub/wiki/How-to:-submit-new-app-to-Hub's-Store">
              <ResponsiveButton
                icon={CirclePlusIcon}
                text="Submit your app"
                variant="outline"
              />
            </ExternalLink>
          </>
        }
      />
      <SuggestedApps />
    </>
  );
}

export default AppStore;
