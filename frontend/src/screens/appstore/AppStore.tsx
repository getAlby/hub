import { CirclePlusIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import ResponsiveButton from "src/components/ResponsiveButton";
import SuggestedApps from "src/components/SuggestedApps";

function AppStore() {
  return (
    <>
      <AppHeader
        title="App Store"
        description="Apps that you can connect your wallet to"
        contentRight={
          <>
            <a
              href="https://github.com/getAlby/hub/wiki/How-to:-submit-new-app-to-Hub's-Store"
              target="_blank"
              rel="noreferrer noopener"
            >
              <ResponsiveButton
                icon={CirclePlusIcon}
                text="Submit your app"
                variant="outline"
              />
            </a>
          </>
        }
      />
      <SuggestedApps />
    </>
  );
}

export default AppStore;
