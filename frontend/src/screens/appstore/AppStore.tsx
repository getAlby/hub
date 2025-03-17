import { CirclePlus } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import SuggestedApps from "src/components/SuggestedApps";
import { Button } from "src/components/ui/button";

function AppStore() {
  return (
    <>
      <AppHeader
        title="App Store"
        contentRight={
          <ExternalLink to="https://github.com/getAlby/hub/wiki/How-to:-submit-new-app-to-Hub's-Store">
            <Button variant="outline">
              <CirclePlus className="h-4 w-4 mr-2" />
              Submit your app
            </Button>
          </ExternalLink>
        }
      />
      <SuggestedApps />
    </>
  );
}

export default AppStore;
