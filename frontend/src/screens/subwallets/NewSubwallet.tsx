import { HelpCircleIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import ResponsiveButton from "src/components/ResponsiveButton";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

export function NewSubwallet() {
  const navigate = useNavigate();
  const [name, setName] = React.useState("");
  const { toast } = useToast();
  const { data: apps } = useApps();
  const { data: info } = useInfo();
  const { data: albyMe, error: albyMeError } = useAlbyMe();

  const [isLoading, setLoading] = React.useState(false);

  if (
    !info ||
    !apps ||
    (info.albyAccountConnected && !albyMe && !albyMeError)
  ) {
    return <Loading />;
  }

  const subwalletApps = apps
    ?.filter(
      (app) => app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID
    )
    .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);

    try {
      if (!albyMe?.subscription.plan_code && subwalletApps?.length >= 3) {
        throw new Error(
          "Max limit reached. Please upgrade to Pro to create more sub-wallets."
        );
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
          app_store_app_id: SUBWALLET_APPSTORE_APP_ID,
        },
      };

      const createAppResponse = await createApp(createAppRequest);

      navigate("/sub-wallets/created", {
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
        title="Create Sub-wallet"
        contentRight={
          <>
            <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/sub-wallets">
              <ResponsiveButton
                icon={HelpCircleIcon}
                text="Help"
                variant="outline"
              />
            </ExternalLink>
          </>
        }
      />
      <form
        onSubmit={handleSubmit}
        className="flex flex-col items-start gap-3 max-w-lg"
      >
        <div className="w-full grid gap-1.5 mb-4">
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
          <p className="text-muted-foreground text-sm">
            Name your friend, family member or coworker
          </p>
        </div>
        <LoadingButton loading={isLoading} type="submit">
          Create Sub-wallet
        </LoadingButton>
      </form>
    </div>
  );
}
