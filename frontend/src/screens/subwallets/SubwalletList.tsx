import {
  HandCoins,
  HelpCircle,
  Landmark,
  TriangleAlert,
  Wallet2,
} from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import UpgradeCard from "src/components/UpgradeCard";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

export function SubwalletList() {
  const navigate = useNavigate();
  const [name, setName] = React.useState("");
  const { data: apps } = useApps();
  const { toast } = useToast();
  const { data: albyMe, error: albyMeError } = useAlbyMe();
  const { data: info } = useInfo();
  const onboardedApps = apps
    ?.filter(
      (app) => app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID
    )
    .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));

  const subwalletAmount = onboardedApps?.reduce(
    (total, app) => total + app.balance,
    0
  );

  const [isLoading, setLoading] = React.useState(false);
  const [showIntro, setShowIntro] = React.useState(true);
  const showForm =
    albyMe?.subscription.plan_code ||
    (onboardedApps && onboardedApps?.length < 3);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);

    try {
      const createAppRequest: CreateAppRequest = {
        name,
        scopes: [
          "get_balance",
          "get_info",
          "list_transactions",
          "lookup_invoice",
          "make_invoice",
          "notifications",
          "pay_invoice",
        ],
        isolated: true,
        metadata: {
          app_store_app_id: SUBWALLET_APPSTORE_APP_ID,
        },
      };

      const createAppResponse = await createApp(createAppRequest);

      navigate(`/sub-wallets/created`, {
        state: createAppResponse,
      });

      toast({ title: "New sub-wallet created for " + name });
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    }
    setLoading(false);
  };

  if (!info || (info.albyAccountConnected && !albyMe && !albyMeError)) {
    // make sure to not render the incorrect component
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Sub-wallets"
        contentRight={
          <>
            {subwalletAmount !== undefined && (
              <Tooltip>
                <TooltipTrigger>
                  <Button variant="outline" className="hidden sm:inline-flex">
                    <Landmark className="w-4 h-4 mr-2" />
                    {new Intl.NumberFormat().format(
                      Math.floor(subwalletAmount / 1000)
                    )}{" "}
                    sats
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  Total amount of assets under management
                </TooltipContent>
              </Tooltip>
            )}
            <ExternalLink to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/app-store/sub-wallet-friends-and-family">
              <Button variant="outline" size="icon">
                <HelpCircle className="w-4 h-4" />
              </Button>
            </ExternalLink>
            <UpgradeDialog>
              <Button variant="premium">Upgrade</Button>
            </UpgradeDialog>
          </>
        }
      />
      {(!showIntro || !!onboardedApps?.length) && (
        <>
          {showForm ? (
            <form
              onSubmit={handleSubmit}
              className="flex flex-col items-start gap-3 max-w-lg"
            >
              <div className="w-full grid gap-1.5">
                <Label htmlFor="name">Sub-wallet name</Label>
                <Input
                  autoFocus
                  type="text"
                  name="name"
                  value={name}
                  id="name"
                  onChange={(e) => setName(e.target.value)}
                  required
                  autoComplete="off"
                />
              </div>
              <LoadingButton loading={isLoading} type="submit">
                Create Sub-wallet
              </LoadingButton>
            </form>
          ) : (
            <UpgradeCard
              title="Need more Sub-wallets?"
              description="Upgrade to Pro to unlock unlimited sub-wallets"
            />
          )}

          {!!onboardedApps?.length && (
            <>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
                {onboardedApps.map((app, index) => (
                  <AppCard key={index} app={app} />
                ))}
              </div>
            </>
          )}
        </>
      )}

      {showIntro && !onboardedApps?.length && (
        <div>
          <div className="flex flex-col gap-6 max-w-screen-md">
            <div className="mb-2">
              <img
                src="/images/illustrations/sub-wallet-dark.svg"
                className="w-72 hidden dark:block"
              />
              <img
                src="/images/illustrations/sub-wallet-light.svg"
                className="w-72 dark:hidden"
              />
            </div>
            <div>
              <div className="flex flex-row gap-3">
                <Wallet2 className="w-6 h-6" />
                <div className="font-medium">
                  Sub-wallets are seperate wallets hosted by your Alby Hub
                </div>
              </div>
              <div className="ml-9 text-muted-foreground text-sm">
                Each sub-wallet has its own balance and can be used as a
                separate wallet that can be connected to Alby Account or any
                app.
              </div>
            </div>
            <div>
              <div className="flex flex-row gap-3">
                <HandCoins className="w-6 h-6" />
                <div className="font-medium">
                  Sub-wallets depend on your Alby Hub spending balance and
                  receive limit
                </div>
              </div>
              <div className="ml-9 text-muted-foreground text-sm">
                Sub-wallets are using your Hubs node liquidity. They can receive
                funds as long as you have enough receive limit in your channels.
              </div>
            </div>
            <div>
              <div className="flex flex-row gap-3">
                <TriangleAlert className="w-6 h-6" />
                <div className="font-medium">
                  Be wary of spending sub-wallets funds
                </div>
              </div>
              <div className="ml-9 text-muted-foreground text-sm">
                If your main balance runs low, funds from isolated wallets can
                be spent by your main wallet balance. You'll receive alerts when
                your spending balance is nearing this point.
              </div>
            </div>
            <div>
              <Button onClick={() => setShowIntro(false)}>
                Create Sub-wallet
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
