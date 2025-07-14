import { CableIcon, CirclePlusIcon, TrashIcon } from "lucide-react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import AlbyConnectionCard from "src/components/connections/AlbyConnectionCard";
import AppCard from "src/components/connections/AppCard";
import { ALBY_ACCOUNT_APP_NAME, SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";
import { useUnusedApps } from "src/hooks/useUnusedApps";

function AppList() {
  const { data: apps } = useApps();
  const { data: info } = useInfo();
  const unusedApps = useUnusedApps();

  if (!apps || !unusedApps || !info) {
    return <Loading />;
  }

  const albyConnection = apps.find((x) => x.name === ALBY_ACCOUNT_APP_NAME);
  const otherApps = apps
    .filter((app) => app.appPubkey !== albyConnection?.appPubkey)
    .filter(
      (app) => app.metadata?.app_store_app_id !== SUBWALLET_APPSTORE_APP_ID
    )
    .sort(
      (a, b) =>
        new Date(b.lastEventAt ?? 0).getTime() -
        new Date(a.lastEventAt ?? 0).getTime()
    );

  return (
    <>
      <AppHeader
        title="Connections"
        contentRight={
          <>
            {!!unusedApps.length && (
              <Link to="/apps/cleanup">
                <ResponsiveButton
                  icon={TrashIcon}
                  text="Cleanup Unused"
                  variant="outline"
                />
              </Link>
            )}
            <Link to="/apps/new">
              <ResponsiveButton icon={CirclePlusIcon} text="Add Connection" />
            </Link>
          </>
        }
      />

      {info.albyAccountConnected && (
        <AlbyConnectionCard connection={albyConnection} />
      )}

      {!otherApps.length && (
        <EmptyState
          icon={CableIcon}
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
