import { Cable, CirclePlus, Trash } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import AlbyConnectionCard from "src/components/connections/AlbyConnectionCard";
import AppCard from "src/components/connections/AppCard";
import { Button } from "src/components/ui/button";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";
import { useIsDesktop } from "src/hooks/useMediaQuery";
import { useUnusedApps } from "src/hooks/useUnusedApps";

const albyConnectionName = "getalby.com";

function AppList() {
  const { data: apps } = useApps();
  const { data: info } = useInfo();
  const isDesktop = useIsDesktop();
  const unusedApps = useUnusedApps();

  if (!apps || !unusedApps || !info) {
    return <Loading />;
  }

  const albyConnection = apps.find((x) => x.name === albyConnectionName);
  const otherApps = apps
    .filter((x) => x.appPubkey !== albyConnection?.appPubkey)
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
          <>
            {!!unusedApps.length && (
              <Link to="/apps/cleanup">
                {isDesktop ? (
                  <Button variant="outline">
                    <Trash className="h-4 w-4 mr-2" />
                    Cleanup Unused
                  </Button>
                ) : (
                  <Button variant="outline" size="icon">
                    <Trash className="w-4 h-4" />
                  </Button>
                )}
              </Link>
            )}
            <Link to="/apps/new">
              {isDesktop ? (
                <Button>
                  <CirclePlus className="h-4 w-4 mr-2" />
                  Add Connection
                </Button>
              ) : (
                <Button size="icon">
                  <CirclePlus className="w-4 h-4" />
                </Button>
              )}
            </Link>
          </>
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
