import { CirclePlusIcon, LayoutGridIcon, Plug2Icon } from "lucide-react";
import React from "react";
import { useSearchParams } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import AppStore from "src/components/connections/AppStore";
import ConnectedApps from "src/components/connections/ConnectedApps";
import ResponsiveButton from "src/components/ResponsiveButton";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "src/components/ui/tabs";

export function Connections() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [tab, setTab] = React.useState(searchParams.get("tab") || "app-store");

  React.useEffect(() => {
    const newTabValue = searchParams.get("tab");
    if (newTabValue) {
      setTab(newTabValue);
      setSearchParams({});
    }
  }, [searchParams, setSearchParams]);

  return (
    <>
      <AppHeader
        title="Connections"
        contentRight={
          <ResponsiveButton icon={CirclePlusIcon} text="Add Connection" />
        }
      />
      <Tabs value={tab} onValueChange={setTab} className="px-2 lg:px-0">
        <TabsList className="mb-2 lg:mb-6">
          <TabsTrigger value="app-store" className="flex gap-2 items-center">
            <LayoutGridIcon className="w-5 h-5" /> App Store
          </TabsTrigger>
          <TabsTrigger
            value="connected-apps"
            className="flex gap-2 items-center"
          >
            <Plug2Icon className="w-5 h-5" /> Connected Apps
          </TabsTrigger>
        </TabsList>
        <TabsContent value="app-store">
          <AppStore />
        </TabsContent>
        <TabsContent value="connected-apps">
          <ConnectedApps />
        </TabsContent>
      </Tabs>
    </>
  );
}
