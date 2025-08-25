import { CirclePlusIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import ResponsiveButton from "src/components/ResponsiveButton";
import SuggestedApps from "src/components/connections/SuggestedApps";

function AppStore() {
  return (
    <>
      <div className="flex flex-col flex-1">
        <div className="flex justify-between items-center">
          <div className="flex-1">
            <h1 className="text-xl lg:text-2xl font-semibold">App Store</h1>
          </div>
          <div className="flex gap-3 h-full">
            <ExternalLink to="https://github.com/getAlby/hub/wiki/How-to:-submit-new-app-to-Hub's-Store">
              <ResponsiveButton
                icon={CirclePlusIcon}
                text="Submit your app"
                variant="outline"
              />
            </ExternalLink>
          </div>
        </div>
      </div>
      <SuggestedApps />
    </>
  );
}

export default AppStore;
