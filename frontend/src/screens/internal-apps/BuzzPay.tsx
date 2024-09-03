import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import Loading from "src/components/Loading";
import { ExternalLinkButton } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { CreateAppRequest, CreateAppResponse } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function BuzzPay() {
  const { data: apps, mutate: reloadApps } = useApps();
  const [creatingApp, setCreatingApp] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const { toast } = useToast();

  if (!apps) {
    return <Loading />;
  }
  const app = apps.find((app) => app.metadata?.app_store_app_id === "buzzpay");

  function createApp() {
    setCreatingApp(true);
    (async () => {
      try {
        const name = "BuzzPay";
        if (apps?.some((existingApp) => existingApp.name === name)) {
          throw new Error("A connection with the same name already exists.");
        }

        const createAppRequest: CreateAppRequest = {
          name,
          scopes: ["get_info", "lookup_invoice", "make_invoice"],
          isolated: true,
          metadata: {
            app_store_app_id: "buzzpay",
          },
        };

        const createAppResponse = await request<CreateAppResponse>(
          "/api/apps",
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify(createAppRequest),
          }
        );

        if (!createAppResponse) {
          throw new Error("no create app response received");
        }

        setConnectionSecret(createAppResponse.pairingUri);

        await reloadApps();

        toast({ title: "BuzzPay app created" });
      } catch (error) {
        handleRequestError(toast, "Failed to create app", error);
      }
      setCreatingApp(false);
    })();
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="BuzzPay"
        description="Receive-only PoS you can safely share with your employees"
      />
      {app && (
        <div className="max-w-lg flex flex-col gap-5">
          <p className="text-muted-foreground">
            Simply click the button below to access your PoS which you can
            instantly receive payments, manage your items, and share your PoS
            with your employees.
          </p>
          <ExternalLinkButton
            to={`https://pos.albylabs.com${connectionSecret && `/#/wallet/${encodeURIComponent(connectionSecret)}/new`}`}
          >
            Go to BuzzPay PoS
          </ExternalLinkButton>
          <AppCard app={app} />
        </div>
      )}
      {!app && (
        <div className="max-w-lg flex flex-col gap-5">
          <p className="text-muted-foreground">
            By creating a new buzzpay app, a read-only wallet connection will be
            created and you will receive a link you can share with your
            employees, on any device.
          </p>
          <LoadingButton loading={creatingApp} onClick={createApp}>
            Create BuzzPay App
          </LoadingButton>
        </div>
      )}
    </div>
  );
}
