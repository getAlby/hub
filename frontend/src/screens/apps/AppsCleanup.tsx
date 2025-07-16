import { SkipForwardIcon, Trash2Icon, TriangleAlertIcon } from "lucide-react";
import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import { Progress } from "src/components/ui/progress";
import { useDeleteApp } from "src/hooks/useDeleteApp";
import { useUnusedApps } from "src/hooks/useUnusedApps";
import { App } from "src/types";

export function AppsCleanup() {
  const unusedApps = useUnusedApps();
  const [appIndex, setAppIndex] = React.useState<number>();
  const [skippedCount, setSkippedCount] = React.useState<number>(0);
  const [deletedCount, setDeletedCount] = React.useState<number>(0);
  const [appsToReview, setAppsToReview] = React.useState<App[]>();
  const { deleteApp } = useDeleteApp();
  React.useEffect(() => {
    if (!unusedApps) {
      return;
    }
    // only allow initializing once
    if (appIndex === undefined) {
      setAppIndex(0);

      const _appsToReview = [...unusedApps];
      _appsToReview.sort((a, b) => {
        return (
          (a.lastEventAt ? new Date(a.lastEventAt).getTime() : 0) -
          (b.lastEventAt ? new Date(b.lastEventAt).getTime() : 0)
        );
      });
      setAppsToReview(_appsToReview);
    }
  }, [appIndex, unusedApps]);

  if (!appsToReview || appIndex === undefined) {
    return <Loading />;
  }

  const currentApp = appsToReview[appIndex];

  return (
    <>
      <AppHeader
        title="Cleanup Unused Apps"
        description="Review apps that haven't been used for 2 months or longer"
      />
      {currentApp && (
        <Alert variant="destructive">
          <AlertTitle className="flex gap-2">
            <TriangleAlertIcon className="h-4 w-4" />
            Warning
          </AlertTitle>
          <AlertDescription>
            Review the app carefully before deleting it, deleted apps cannot be
            recovered.
          </AlertDescription>
        </Alert>
      )}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
        <div className="lg:col-span-3 flex flex-col gap-3">
          {currentApp && (
            <>
              <div className="w-full h-full">
                <AppCard
                  app={currentApp}
                  actions={
                    <>
                      <Button
                        onClick={() => {
                          setAppIndex(appIndex + 1);
                          setSkippedCount((current) => current + 1);
                        }}
                      >
                        <SkipForwardIcon className="h-4 w-4 mr-2" />
                        Skip
                      </Button>
                      <Button
                        variant="destructive"
                        onClick={() => {
                          deleteApp(currentApp.appPubkey);
                          setAppIndex(appIndex + 1);
                          setDeletedCount((current) => current + 1);
                        }}
                      >
                        <Trash2Icon className="h-4 w-4 mr-2" />
                        Delete
                      </Button>
                    </>
                  }
                  readonly
                />
              </div>
            </>
          )}
          {!currentApp && (
            <>
              <div>No more unused apps to review.</div>
              <div>
                <LinkButton to="/apps">Back to overview</LinkButton>
              </div>
            </>
          )}
        </div>
        <div className="lg:col-span-2">
          <Card className="w-full">
            <CardHeader>
              <CardTitle>Progress</CardTitle>
              <CardDescription>
                {appIndex} of {appsToReview.length} unused apps reviewed (
                {skippedCount} skipped, {deletedCount} deleted)
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <Progress value={(appIndex / appsToReview.length) * 100} />
            </CardFooter>
          </Card>
        </div>
      </div>
    </>
  );
}
