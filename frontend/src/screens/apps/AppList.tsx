import { Cable, CirclePlus } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import AlbyConnectionCard from "src/components/connections/AlbyConnectionCard";
import AppCard from "src/components/connections/AppCard";
import { Button } from "src/components/ui/button";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";

const albyConnectionName = "getalby.com";

function AppList() {
  const { data: apps } = useApps();
  const { data: info } = useInfo();

  if (!apps || !info) {
    return <Loading />;
  }

  const albyConnection = apps.find((x) => x.name === albyConnectionName);
  const otherApps = apps
    .filter((x) => x.nostrPubkey !== albyConnection?.nostrPubkey)
    .sort(
      (a, b) =>
        new Date(b.lastEventAt ?? 0).getTime() -
        new Date(a.lastEventAt ?? 0).getTime()
    );

  return (
    <>
      <AppHeader
        title="Connections"
        description="Apps that you connected to already"
        contentRight={
          <Link to="/apps/new">
            <Button>
              <CirclePlus className="h-4 w-4 mr-2" />
              Add Connection
            </Button>
          </Link>
        }
      />

      {info.albyAccountConnected && (
        <AlbyConnectionCard connection={albyConnection} />
      )}

      {!otherApps.length && (
        <EmptyState
          icon={Cable}
          title="Connect Your First App"
          description="Connect your app of choice, fine-tune permissions and enjoy a seamless and secure wallet experience."
          buttonText="See Recommended Apps"
          buttonLink="/appstore"
        />
      )}

      {otherApps.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
          {otherApps.map((app, index) => (
            <AppCard key={index} app={app} />
          ))}
        </div>
      )}
    </>
  );
}

export default AppList;
