import { Button } from "src/components/ui/button";
import SuggestedApps from "src/components/SuggestedApps";
import { CirclePlus } from "lucide-react";
import AppHeader2 from "src/components/AppHeader2";

function AppsList() {
  return (
    <>
      <AppHeader2
        title="Apps"
        description="Apps that you can connect your wallet into"
        contentRight={
          <>
            <Button variant="secondary">How to connect to apps?</Button>
            <Button variant="outline">
              <CirclePlus className="h-4 w-4 mr-2" />
              Submit your app
            </Button>
          </>
        }
      />
      <h2 className="text-md font-medium md:text-xl">Featured</h2>
      TBD
      <h2 className="text-md font-medium md:text-xl">All apps</h2>
      <SuggestedApps />
    </>
  );
}

export default AppsList;
