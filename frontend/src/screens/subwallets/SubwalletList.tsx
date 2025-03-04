import { HandCoins, HelpCircle, TriangleAlert, Wallet2 } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

export function SubwalletList() {
  const navigate = useNavigate();
  const [name, setName] = React.useState("");
  const { data: apps } = useApps();
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);
  const [showIntro, setShowIntro] = React.useState(true);

  const onboardedApps = apps
    ?.filter((app) => app.metadata?.app_store_app_id === "uncle-jim")
    .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);

    try {
      if (apps?.some((existingApp) => existingApp.name === name)) {
        throw new Error("A connection with the same name already exists.");
      }

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
          app_store_app_id: "uncle-jim",
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

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Sub-wallets"
        description="Create personal spaces for your bitcoin with sub-wallets — keep funds organized for yourself, family and friends"
        contentRight={
          <ExternalLink to="https://guides.getalby.com/user-guide/alby-account-and-browser-extension/alby-hub/app-store/sub-wallet-friends-and-family">
            <Button variant="outline" size="icon">
              <HelpCircle className="w-4 h-4" />
            </Button>
          </ExternalLink>
        }
      />
      {(!showIntro || !!onboardedApps?.length) && (
        <>
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
