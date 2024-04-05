import { Cable, CirclePlus } from "lucide-react";
import { Link } from "react-router-dom";
import AppCard from "src/components/AppCard";
import AppHeader from "src/components/AppHeader";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";

function AppList() {
  const { data: apps } = useApps();
  const { data: info } = useInfo();

  if (!apps || !info) {
    return <Loading />;
  }

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

      {!apps.length && (
        <EmptyState
          icon={<Cable />}
          title="You have no connections, yet"
          description="Create your first one by checking out our recommended apps"
          buttonText="See Recommended Apps"
          buttonLink="/appstore"
        />
      )}

      {apps.length > 0 && (
        <div className="grid sm:grid-cols-1 md:grid-cols-2 gap-4">
          {apps.map((app, index) => (
            <AppCard key={index} app={app} />
          ))}
        </div>
      )}
    </>
  );
}

export default AppList;
