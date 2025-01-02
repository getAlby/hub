import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

import alby from "src/assets/suggested-apps/alby.png";
import hrf from "src/assets/zapplanner/hrf.png";
import opensats from "src/assets/zapplanner/opensats.png";

type Recipient = {
  name: string;
  description: string;
  lightningAddress: string;
  logo?: string;
};

const recipients: Recipient[] = [
  {
    name: "Alby",
    logo: alby,
    description:
      "Support the open source development of Alby Hub, Alby Go, Alby developer tools, NWC protocol development, and more.",
    lightningAddress: "hello@getalby.com",
  },
  {
    name: "HRF",
    description:
      "We collaborate with transformative activists to develop innovative solutions that bring the world together in the fight against tyranny.",
    lightningAddress: "hrf@btcpay.hrf.org",
    logo: hrf,
  },
  {
    name: "OpenSats",
    description:
      "Help us to provide sustainable funding for free and open-source contributors working on freedom tech and projects that help bitcoin flourish.",
    lightningAddress: "opensats@btcpay0.voltageapp.io",
    logo: opensats,
  },
];

export function ZapPlanner() {
  const { data: apps, mutate: reloadApps } = useApps();
  const { toast } = useToast();
  const [submittingFor, setSubmittingFor] = React.useState("");
  const [customName, setCustomName] = React.useState("");
  const [customLightningAddress, setCustomLightningAddress] =
    React.useState("");

  const handleSubmit = async (name: string, lightningAddress: string) => {
    setSubmittingFor(lightningAddress);

    try {
      if (apps?.some((existingApp) => existingApp.name === name)) {
        throw new Error("A connection with the same name already exists.");
      }

      const amountSats = 5000;

      const createAppRequest: CreateAppRequest = {
        name: `ZapPlanner - ${name}`,
        scopes: [
          "get_balance",
          "get_info",
          "list_transactions",
          "lookup_invoice",
          "make_invoice",
          "notifications",
          "pay_invoice",
        ],
        budgetRenewal: "monthly",
        maxAmount: Math.floor(amountSats * 1.01) + 10 /* with fee reserve */,
        isolated: false,
        metadata: {
          app_store_app_id: "zapplanner",
          recipient_lightning_address: lightningAddress,
        },
      };

      const createAppResponse = await createApp(createAppRequest);
      /*toast({
        title: "Created subscription",
        description: "The first payment is scheduled for 1st of October.",
      });*/
      const comment = encodeURIComponent("ZapPlanner payment from Alby Hub"); // TODO: allow customization
      const payerData = ""; //encodeURIComponent(JSON.stringify({}));
      // TODO: consider fiat conversion
      // TODO: consider using the ZapPlanner API to make the connection, then the subscription ID can be saved in the app metadata
      window.open(
        `https://zapplanner.albylabs.com/confirm?amount=${amountSats}&recipient=${lightningAddress}&timeframe=31%20days&comment=${comment}&payerdata=${payerData}&nwcUrl=${encodeURIComponent(createAppResponse.pairingUri)}`,
        "_blank"
      );

      reloadApps();
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    }
    setSubmittingFor("");
  };

  const zapplannerApps = apps?.filter(
    (app) => app.metadata?.app_store_app_id === "zapplanner"
  );

  return (
    <div className="grid gap-5">
      <AppHeader
        title="ZapPlanner"
        description="Schedule automatic recurring lightning payments"
      />
      {/* TODO: add a dialog so a comment can be passed */}
      <div className="grid grid-cols-3 gap-3">
        {recipients.map((recipient) => (
          <Card key={recipient.lightningAddress}>
            <CardHeader className="h-32">
              <div className="flex flex-row items-center gap-3">
                <img src={recipient.logo} className="rounded-lg w-10 h-10" />
                <CardTitle>{recipient.name}</CardTitle>
              </div>
              <CardDescription>{recipient.description}</CardDescription>
            </CardHeader>
            <CardFooter className="flex justify-end">
              <LoadingButton
                disabled={!!submittingFor}
                size="sm"
                loading={submittingFor === recipient.lightningAddress}
                onClick={() => {
                  handleSubmit(recipient.name, recipient.lightningAddress);
                }}
              >
                Support with 5000 sats / month
              </LoadingButton>
            </CardFooter>
          </Card>
        ))}
        <Card>
          <CardHeader className="h-32">
            <CardTitle>Custom</CardTitle>
            <CardDescription className="flex flex-col gap-2">
              <Input
                value={customName}
                onChange={(e) => setCustomName(e.target.value)}
                placeholder={"Name"}
              />
              <Input
                value={customLightningAddress}
                onChange={(e) => setCustomLightningAddress(e.target.value)}
                placeholder={"lightning@address.com"}
              />
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-end mt-2">
            <LoadingButton
              disabled={
                !!submittingFor || !customName || !customLightningAddress
              }
              loading={!!submittingFor}
              onClick={() => {
                handleSubmit(customName, customLightningAddress);
              }}
            >
              Support with 5000 sats / month
            </LoadingButton>
          </CardFooter>
        </Card>
      </div>

      {!!zapplannerApps?.length && (
        <>
          <h2 className="font-semibold text-xl">Recurring Payments</h2>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
            {zapplannerApps.map((app, index) => (
              <AppCard key={index} app={app} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
