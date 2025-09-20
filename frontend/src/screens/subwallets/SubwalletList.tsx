import {
  CirclePlusIcon,
  HelpCircle,
  InfoIcon,
  ShieldCheckIcon,
  SparklesIcon,
  TriangleAlert,
  TriangleAlertIcon,
} from "lucide-react";
import { useRef, useState } from "react";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import { CustomPagination } from "src/components/CustomPagination";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { LIST_APPS_LIMIT, SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { SubwalletIntro } from "src/screens/subwallets/SubwalletIntro";

export function SubwalletList() {
  const { data: info } = useInfo();
  const [page, setPage] = useState(1);
  const appsListRef = useRef<HTMLDivElement>(null);
  const { data: appsData } = useApps(
    undefined,
    page,
    {
      appStoreAppId: SUBWALLET_APPSTORE_APP_ID,
    },
    "created_at"
  );
  const { data: albyMe, error: albyMeError } = useAlbyMe();
  const { data: balances } = useBalances();

  const handlePageChange = (page: number) => {
    setPage(page);
    appsListRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  };

  if (
    !info ||
    !appsData ||
    !balances ||
    (info.albyAccountConnected && !albyMe && !albyMeError)
  ) {
    return <Loading />;
  }

  const subwalletApps = appsData.apps;

  if (!subwalletApps.length) {
    return <SubwalletIntro />;
  }

  const subwalletTotalAmount =
    subwalletApps.reduce((total, app) => total + app.balance, 0) || 0;
  const isSufficientlyBacked =
    subwalletTotalAmount <= balances.lightning.totalSpendable;

  return (
    <div className="grid gap-4">
      <AppHeader
        title="Sub-wallets"
        description="Create sub-wallets for yourself, friends, family or coworkers"
        contentRight={
          <>
            <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/sub-wallets">
              <Button variant="outline" size="icon">
                <HelpCircle className="size-4" />
              </Button>
            </ExternalLink>
            {!albyMe?.subscription.plan_code && subwalletApps?.length >= 3 ? (
              <UpgradeDialog>
                <ResponsiveButton icon={CirclePlusIcon} text="New Sub-wallet" />
              </UpgradeDialog>
            ) : (
              <Link to="/sub-wallets/new">
                <ResponsiveButton icon={CirclePlusIcon} text="New Sub-wallet" />
              </Link>
            )}
          </>
        }
      />

      {!albyMe?.subscription.plan_code && subwalletApps.length >= 3 && (
        <>
          <Alert>
            <InfoIcon />
            <AlertTitle>Need more Sub-wallets?</AlertTitle>
            <AlertDescription className="flex flex-row gap-3">
              <p className="grow">
                Upgrade your subscription plan to Pro unlock unlimited number of
                Sub-wallets.
              </p>
              <UpgradeDialog>
                <Button>
                  <SparklesIcon />
                  Upgrade
                </Button>
              </UpgradeDialog>
            </AlertDescription>
          </Alert>
        </>
      )}

      {!isSufficientlyBacked && (
        <Alert variant="warning">
          <TriangleAlert />
          <AlertTitle>
            Sub-wallets you manage are insufficiently backed
          </AlertTitle>
          <AlertDescription className="flex flex-row gap-3">
            There's not enough bitcoin in your spending balance to honor all
            balances of sub-wallets under your management. Increase spending
            capacity by opening a channel or review your channel statuses to
            back them up again.
            <Link to="/wallet/receive">
              <Button variant="secondary">Deposit Bitcoin</Button>
            </Link>
          </AlertDescription>
        </Alert>
      )}

      <div className="flex flex-col sm:flex-row flex-wrap gap-4 slashed-zero">
        <Card className="flex flex-1 flex-col">
          <CardHeader className="pb-2">
            <CardTitle className="text-lg">
              Total Balance of Sub-wallets
            </CardTitle>
          </CardHeader>
          <CardContent className="grow">
            <div className="mb-1">
              <span className="text-2xl font-medium balance sensitive">
                {new Intl.NumberFormat().format(
                  Math.floor(subwalletTotalAmount / 1000)
                )}{" "}
                sats
              </span>
            </div>
            <FormattedFiatAmount amount={subwalletTotalAmount / 1000} />
          </CardContent>
        </Card>
        <Card className="flex flex-1 flex-col">
          <CardHeader className="pb-2">
            <CardTitle className="text-lg">Number of Sub-wallets</CardTitle>
          </CardHeader>
          <CardContent className="grow flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <span className="text-2xl font-medium">
                {subwalletApps.length} /{" "}
                {albyMe?.subscription.plan_code ? "∞" : 3}
              </span>
              {isSufficientlyBacked ? (
                <div className="flex items-center text-positive-foreground text-sm">
                  <ShieldCheckIcon className="size-4 mr-2" />
                  <span className="text-sm font-medium">Fully backed</span>
                </div>
              ) : (
                <div className="flex items-center text-warning-foreground text-sm">
                  <TriangleAlertIcon className="size-4 mr-2" />
                  <span className="text-sm font-medium">
                    Insufficiently backed
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
      <div className="mt-8">
        <h3 className="font-semibold text-2xl mb-4">Managed Sub-wallets</h3>
        <div
          ref={appsListRef}
          className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list"
        >
          {subwalletApps.map((app, index) => (
            <AppCard key={index} app={app} />
          ))}
        </div>
      </div>

      <CustomPagination
        limit={LIST_APPS_LIMIT}
        totalCount={appsData.totalCount}
        page={page}
        handlePageChange={handlePageChange}
      />
    </div>
  );
}
