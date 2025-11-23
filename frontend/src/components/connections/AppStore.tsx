import { CirclePlusIcon } from "lucide-react";
import ResponsiveExternalLinkButton from "src/components/ResponsiveExternalLinkButton";
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
            <ResponsiveExternalLinkButton
              icon={CirclePlusIcon}
              text="Submit your app"
              variant="outline"
              to="https://github.com/getAlby/hub/wiki/How-to:-submit-new-app-to-Hub's-Store"
            />
          </div>
        </div>
      </div>
      <SuggestedApps />
    </>
  );
}

export default AppStore;
