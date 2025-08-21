import { CableIcon, TrashIcon } from "lucide-react";
import { useRef, useState } from "react";
import { Link } from "react-router-dom";
import { CustomPagination } from "src/components/CustomPagination";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import AlbyConnectionCard from "src/components/connections/AlbyConnectionCard";
import AppCard from "src/components/connections/AppCard";
import {
  ALBY_ACCOUNT_APP_NAME,
  LIST_APPS_LIMIT,
  SUBWALLET_APPSTORE_APP_ID,
} from "src/constants";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";
import { useUnusedApps } from "src/hooks/useUnusedApps";
import { ListAppsResponse } from "src/types";

// display previous page while next page is loading
let prevAppsData: ListAppsResponse | undefined;

function ConnectedApps() {
  const { data: info } = useInfo();
  const [page, setPage] = useState(1);
  const { data: appsData } = useApps(LIST_APPS_LIMIT, page);
  const appsListRef = useRef<HTMLDivElement>(null);
  const handlePageChange = (page: number) => {
    setPage(page);
    appsListRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  };

  const unusedApps = useUnusedApps();

  if ((!prevAppsData && !appsData) || !unusedApps || !info) {
    return <Loading />;
  }
  if (appsData) {
    prevAppsData = appsData;
  }
  const { apps, totalCount } = appsData || prevAppsData!;

  const otherApps = apps
    .filter((app) => app.name !== ALBY_ACCOUNT_APP_NAME)
    .filter(
      (app) => app.metadata?.app_store_app_id !== SUBWALLET_APPSTORE_APP_ID
    );

  return (
    <>
      <div className="flex flex-col flex-1">
        <div className="flex justify-between items-center">
          <div className="flex-1">
            <h1 className="text-xl lg:text-2xl font-semibold">
              Connected Apps
            </h1>
          </div>
          <div className="flex gap-3 h-full">
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
              {/* <Link to="/apps/new">
              <ResponsiveButton icon={CirclePlusIcon} text="Add Connection" />
            </Link> */}
            </>
          </div>
        </div>
      </div>

      {info.albyAccountConnected && (
        <div className="mt-6">
          <AlbyConnectionCard />
        </div>
      )}

      <div className="mt-6" />

      {!otherApps.length && (
        <EmptyState
          icon={CableIcon}
          title="Connect Your First App"
          description="Connect your app of choice, fine-tune permissions and enjoy a seamless and secure wallet experience."
          buttonText="See Recommended Apps"
          buttonLink="/apps?tab=app-store"
        />
      )}

      {otherApps.length > 0 && (
        <div
          ref={appsListRef}
          className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list"
        >
          {otherApps.map((app, index) => (
            <AppCard key={index} app={app} />
          ))}
        </div>
      )}

      <CustomPagination
        limit={LIST_APPS_LIMIT}
        totalCount={totalCount}
        page={page}
        handlePageChange={handlePageChange}
      />
    </>
  );
}

export default ConnectedApps;
