import { Button } from "src/components/ui/button";
import SuggestedApps from "src/components/SuggestedApps";
import { CirclePlus } from "lucide-react";

function AppsList() {
  return (
    <>
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold md:text-2xl">Apps</h1>
        <div className="flex flex-row gap-3">
          <Button variant="secondary">How to connect to apps?</Button>
          <Button variant="outline">
            <CirclePlus className="h-4 w-4 mr-2" />
            Submit your app
          </Button>
        </div>
      </div>
      <h2 className="text-md font-medium md:text-xl">Featured</h2>
      TBD
      <h2 className="text-md font-medium md:text-xl">All apps</h2>
      <SuggestedApps />
    </>
  );
}

export default AppsList;
