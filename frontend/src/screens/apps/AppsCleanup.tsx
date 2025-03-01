import dayjs from "dayjs";
import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import Loading from "src/components/Loading";
import { Button, LinkButton } from "src/components/ui/button";
import { useApps } from "src/hooks/useApps";
import { useDeleteApp } from "src/hooks/useDeleteApp";
import { App } from "src/types";

export function AppsCleanup() {
  const { data: apps } = useApps();
  const [appIndex, setAppIndex] = React.useState<number>();
  const [skippedCount, setSkippedCount] = React.useState<number>(0);
  const [deletedCount, setDeletedCount] = React.useState<number>(0);
  const [appsToReview, setAppsToReview] = React.useState<App[]>();
  const { deleteApp } = useDeleteApp();
  React.useEffect(() => {
    if (!apps) {
      return;
    }
    // only allow initializing once
    if (appIndex === undefined) {
      setAppIndex(0);
      const oldDate = dayjs().subtract(1, "month");
      const _appsToReview = apps.filter(
        (app) => !app.lastEventAt || dayjs(app.lastEventAt).isBefore(oldDate)
      );
      _appsToReview.sort((a, b) => {
        return (
          (a.lastEventAt ? new Date(a.lastEventAt).getTime() : 0) -
          (b.lastEventAt ? new Date(b.lastEventAt).getTime() : 0)
        );
      });
      setAppsToReview(_appsToReview);
    }
  }, [appIndex, apps]);

  if (!appsToReview || appIndex === undefined) {
    return <Loading />;
  }

  const currentApp = appsToReview[appIndex];

  return (
    <div className="max-w-lg flex flex-col gap-4 items-center">
      <div className="w-full">
        <AppHeader
          title="App Cleanup"
          description="Quickly remove unused apps"
        />
      </div>
      <p className="text-muted-foreground">
        Review the app carefully before deleting it. Deleted apps cannot be
        recovered. Take care before deleting{" "}
        <span className="font-semibold">Friends & Family</span> apps.
      </p>
      {currentApp && (
        <>
          <p className="font-mono">
            {appIndex + 1} / {appsToReview.length} unused apps to review,{" "}
            {skippedCount} skipped, {deletedCount} deleted
          </p>

          <div className="w-full">
            <AppCard
              app={currentApp}
              actions={
                <div className="flex gap-2">
                  <Button
                    variant="destructive"
                    className="w-full"
                    onClick={() => {
                      deleteApp(currentApp.appPubkey);
                      setAppIndex(appIndex + 1);
                      setDeletedCount((current) => current + 1);
                    }}
                  >
                    Delete
                  </Button>
                  <Button
                    variant="secondary"
                    className="w-full"
                    onClick={() => {
                      setAppIndex(appIndex + 1);
                      setSkippedCount((current) => current + 1);
                    }}
                  >
                    Skip
                  </Button>
                </div>
              }
            />
          </div>
        </>
      )}
      {!currentApp && (
        <>
          <p className="my-8">🎉 All apps reviewed!</p>
          <LinkButton to="/" className="w-full">
            Home
          </LinkButton>
          <Button
            variant="secondary"
            className="w-full"
            onClick={() => window.location.reload()}
          >
            Restart
          </Button>
        </>
      )}
    </div>
  );
}
