import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import Loading from "src/components/Loading";
import { ExternalLinkButton } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import {
  App,
  AppPermissions,
  CreateAppRequest,
  CreateAppResponse,
  UpdateAppRequest,
} from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function BuzzPay() {
  const { data: apps, mutate: reloadApps } = useApps();
  const [creatingApp, setCreatingApp] = React.useState(false);
  const { toast } = useToast();

  if (!apps) {
    return <Loading />;
  }
  const app = apps.find(
    (app) =>
      app.metadata?.app_store_app_id === "buzzpay" &&
      app.metadata.connection_secret
  );

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
          isolated: false,
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

        const app = await request<App>(
          `/api/apps/${createAppResponse.pairingPublicKey}`
        );

        if (!app) {
          throw new Error("failed to fetch buzzpay app");
        }

        const permissions: AppPermissions = {
          scopes: app.scopes,
          maxAmount: app.maxAmount,
          budgetRenewal: app.budgetRenewal,
          expiresAt: app.expiresAt ? new Date(app.expiresAt) : undefined,
          isolated: app.isolated,
        };

        // TODO: should be able to partially update app rather than having to pass everything
        // we are only updating the metadata
        const updateAppRequest: UpdateAppRequest = {
          name,
          scopes: Array.from(permissions.scopes),
          budgetRenewal: permissions.budgetRenewal,
          expiresAt: permissions.expiresAt?.toISOString(),
          maxAmount: permissions.maxAmount,
          metadata: {
            ...app.metadata,
            // read-only connection secret
            connection_secret: createAppResponse.pairingUri,
          },
        };

        await request(`/api/apps/${app.nostrPubkey}`, {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(updateAppRequest),
        });
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
        <div>
          <AppCard app={app} />
          <ExternalLinkButton
            className="mt-4"
            to={`https://pos.albylabs.com/#/wallet/${encodeURIComponent(app.metadata?.connection_secret as string)}/new`}
          >
            Go to BuzzPay PoS
          </ExternalLinkButton>
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
