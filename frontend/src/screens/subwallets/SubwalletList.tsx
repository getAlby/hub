import {
  CirclePlusIcon,
  HelpCircle,
  InfoIcon,
  ShieldCheckIcon,
  SparklesIcon,
  TriangleAlertIcon,
} from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import ExternalLink from "src/components/ExternalLink";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";
import { SubwalletIntro } from "src/screens/subwallets/SubwalletIntro";

export function SubwalletList() {
  const { data: info } = useInfo();
  const { data: apps } = useApps();
  const { data: albyMe, error: albyMeError } = useAlbyMe();
  const { data: balances } = useBalances();
  const navigate = useNavigate();

  if (
    !info ||
    !apps ||
    !balances ||
    (info.albyAccountConnected && !albyMe && !albyMeError)
  ) {
    return <Loading />;
  }

  const subwalletApps = apps
    ?.filter(
      (app) => app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID
    )
    .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));

  if (!subwalletApps?.length) {
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
            <ExternalLink to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/app-store/sub-wallet-friends-and-family">
              <Button variant="outline" size="icon">
                <HelpCircle className="w-4 h-4" />
              </Button>
            </ExternalLink>
            <ResponsiveButton
              icon={CirclePlusIcon}
              text="New Sub-wallet"
              disabled={
                !albyMe?.subscription.plan_code && subwalletApps?.length >= 3
              }
              onClick={() => navigate("/sub-wallets/new")}
            />
          </>
        }
      />
      {!isSufficientlyBacked && (
        <Alert className="text-warning-foreground border-warning-foreground flex items-center gap-4">
          <div className="flex gap-3">
            <InfoIcon className="!text-warning-foreground h-4 w-4 flex-shrink-0" />
            <div>
              <AlertTitle>
                Sub-wallets you manage are insufficiently backed
              </AlertTitle>
              <AlertDescription>
                There's not enough bitcoin in your spending balance to honor all
                balances of sub-wallets under your management. Increase spending
                capacity by opening a channel or review your channel statuses to
                back them up again.
              </AlertDescription>
            </div>
          </div>
          <Link to="/channels">
            <Button>Go to Node</Button>
          </Link>
        </Alert>
      )}
      <div
        className={cn("flex flex-col sm:flex-row flex-wrap gap-4 slashed-zero")}
      >
        <Card className="flex flex-1 flex-col">
          <CardHeader className="pb-2 space-y-0">
            <CardTitle className="text-lg">Assets Under Management</CardTitle>
            <CardDescription className="mt-0">
              Total amount of assets under management
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-grow">
            <div className="mt-4 mb-1">
              <span className="text-2xl font-medium balance sensitive">
                {new Intl.NumberFormat().format(
                  Math.floor(subwalletTotalAmount / 1000)
                )}{" "}
                sats
              </span>
            </div>
            <FormattedFiatAmount
              amount={subwalletTotalAmount / 1000}
              className="mb-1"
            />
          </CardContent>
        </Card>
        <Card className="flex flex-1 flex-col">
          <CardHeader className="pb-2 space-y-0">
            <CardTitle className="text-lg">Active Sub-wallets</CardTitle>
            <CardDescription className="mt-0">
              Number of Sub-wallets backed by your node funds
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-grow flex flex-col gap-4">
            <div className="flex flex-row items-center justify-between mt-4">
              <span className="text-4xl font-medium balance sensitive mb-1">
                {subwalletApps.length} /{" "}
                {albyMe?.subscription.plan_code ? "∞" : 3}
              </span>
              {isSufficientlyBacked ? (
                <div className="flex items-center text-positive-foreground">
                  <ShieldCheckIcon className="w-4 h-4 mr-2" />
                  <span className="text-sm font-medium">Fully backed</span>
                </div>
              ) : (
                <div className="flex items-center text-warning-foreground">
                  <TriangleAlertIcon className="w-4 h-4 mr-2" />
                  <span className="text-sm font-medium">
                    Insufficiently backed
                  </span>
                </div>
              )}
            </div>
            {!albyMe?.subscription.plan_code && (
              <>
                <div className="flex gap-3">
                  <InfoIcon className="w-6 h-6" />
                  <div>
                    <h4 className="font-medium mb-1">
                      Upgrade to unlock unlimited Sub-wallets
                    </h4>
                    <p className="text-sm text-muted-foreground">
                      Upgrade your plan to Pro to be able to create unlimited
                      number of Sub-wallets
                    </p>
                  </div>
                </div>

                <div className="flex justify-end">
                  <UpgradeDialog>
                    <Button variant="secondary">
                      <SparklesIcon className="w-4 h-4 mr-2" />
                      Upgrade Your Plan
                    </Button>
                  </UpgradeDialog>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
      <div className="mt-8">
        <h3 className="font-semibold text-2xl mb-4">Managed Sub-wallets</h3>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
          {subwalletApps.map((app, index) => (
            <AppCard key={index} app={app} />
          ))}
        </div>
      </div>
    </div>
  );
}
