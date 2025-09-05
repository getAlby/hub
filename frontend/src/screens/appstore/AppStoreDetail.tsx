import React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AboutAppCard } from "src/components/connections/AboutAppCard";
import { AppLinksCard } from "src/components/connections/AppLinksCard";
import { AppStoreDetailHeader } from "src/components/connections/AppStoreDetailHeader";
import {
  AppStoreApp,
  appStoreApps,
} from "src/components/connections/SuggestedAppData";
import Loading from "src/components/Loading";
import { useAppsForAppStoreApp } from "src/hooks/useApps";

export function AppStoreDetail() {
  const { appStoreId } = useParams() as { appStoreId: string };
  const appStoreApp = appStoreApps.find((x) => x.id === appStoreId);
  const navigate = useNavigate();

  if (!appStoreApp) {
    navigate("/apps?tab=app-store");
    return null;
  }

  // Redirect internal apps to their dedicated pages
  if (appStoreApp.internal) {
    navigate(`/internal-apps/${appStoreId}`);
    return null;
  }

  return (
    <AppStoreDetailInternal appStoreId={appStoreId} appStoreApp={appStoreApp} />
  );
}

function AppStoreDetailInternal({
  appStoreApp,
  appStoreId,
}: {
  appStoreApp: AppStoreApp;
  appStoreId: string;
}) {
  const connectedApps = useAppsForAppStoreApp(appStoreApp);
  const navigate = useNavigate();

  React.useEffect(() => {
    if (connectedApps && connectedApps.length > 0) {
      navigate(`/apps/${connectedApps[0].id}`, {
        replace: true,
      });
    }
  }, [connectedApps, navigate]);

  if (!connectedApps) {
    return <Loading />;
  }

  return (
    <div className="grid gap-2">
      <AppStoreDetailHeader appStoreApp={appStoreApp} />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
        <AboutAppCard appStoreApp={appStoreApp} />
        <AppLinksCard appStoreApp={appStoreApp} />
        {/* <Card>
          <CardHeader>
            <CardTitle className="text-2xl">How to Connect</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            {appStoreApp.guide || (
              <ul className="list-inside list-decimal">
                <li>Install the app</li>
                <li>
                  Click{" "}
                  <Link to={`/apps/new?app=${appStoreId}`}>
                    <Button variant="link" className="px-0">
                      Connect to {appStoreApp.title}
                    </Button>
                  </Link>
                </li>
                <li>
                  Open the Alby Go app on your mobile and scan the QR code
                </li>
              </ul>
            )}
          </CardContent>
        </Card> */}
      </div>
    </div>
  );
}
