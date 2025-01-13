import { CirclePlus } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import SuggestedApps from "src/components/SuggestedApps";
import { Button } from "src/components/ui/button";

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
              <Button variant="outline">
                <CirclePlus className="h-4 w-4 mr-2" />
                Submit your app
              </Button>
            </a>
          </>
        }
      />
      <SuggestedApps />
    </>
  );
}

export default AppStore;
