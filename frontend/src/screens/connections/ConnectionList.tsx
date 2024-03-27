import { Link } from "react-router-dom";

import { CirclePlus } from "lucide-react";
import Loading from "src/components/Loading";

import AppCard from "src/components/AppCard";
import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";

function ConnectionList() {
  const { data: apps, mutate: mutateApps } = useApps();
  const { data: info } = useInfo();

  if (!apps || !info) {
    return <Loading />;
  }

  const handleDeleteApp = (nostrPubkey: string) => {
    const updatedApps = apps.filter((app) => app.nostrPubkey !== nostrPubkey);
    mutateApps(updatedApps, false);
  };

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
        <>
          <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm">
            <div className="flex flex-col items-center gap-1 text-center">
              <h3 className="text-2xl font-bold tracking-tight">
                You have no connections, yet
              </h3>
              <p className="text-sm text-muted-foreground">
                Create your first one by checking out our recommended apps
              </p>
              <Link to="/apps">
                <Button className="mt-4">See recommended apps</Button>
              </Link>
            </div>
          </div>
        </>
      )}

      {apps.length > 0 && (
        <div className="grid sm:grid-cols-1 md:grid-cols-2 gap-4">
          {apps.map((app, index) => (
            <AppCard key={index} app={app} onDelete={handleDeleteApp} />
          ))}
        </div>
      )}
    </>
  );
}

export default ConnectionList;
