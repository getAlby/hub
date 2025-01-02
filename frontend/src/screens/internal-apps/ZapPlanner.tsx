import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

import { PlusCircle } from "lucide-react";
import alby from "src/assets/suggested-apps/alby.png";
import hrf from "src/assets/zapplanner/hrf.png";
import opensats from "src/assets/zapplanner/opensats.png";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";

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

  const [open, setOpen] = React.useState(false);
  const [isSubmitting, setSubmittingFor] = React.useState("");
  const [customName, setCustomName] = React.useState("");
  const [customLightningAddress, setCustomLightningAddress] =
    React.useState("");

  const handleSubmit = async () => {
    try {
      if (apps?.some((existingApp) => existingApp.name === customName)) {
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
          recipient_lightning_address: customLightningAddress,
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
        `https://zapplanner.albylabs.com/confirm?amount=${amountSats}&recipient=${customLightningAddress}&timeframe=31%20days&comment=${comment}&payerdata=${payerData}&nwcUrl=${encodeURIComponent(createAppResponse.pairingUri)}`,
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
        contentRight={
          <>
            <Dialog open={open} onOpenChange={setOpen}>
              <DialogTrigger asChild>
                <Button>
                  <PlusCircle className="h-4 w-4 mr-2" />
                  New Recurring Payment
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[600px]">
                <DialogHeader>
                  <DialogTitle>New Recurring Payment</DialogTitle>
                  <DialogDescription>
                    For advanced options go to{" "}
                    <ExternalLink
                      className="underline"
                      to="https://zapplanner.albylabs.com"
                    >
                      zapplanner.albylabs.com
                    </ExternalLink>{" "}
                    and create your recurring payment there.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="name" className="text-right">
                      Name
                    </Label>
                    <Input
                      id="name"
                      value={customName}
                      className="col-span-3"
                    />
                  </div>
                  <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="name" className="text-right">
                      Receiver
                    </Label>
                    <Input
                      id="receiver"
                      value={customLightningAddress}
                      className="col-span-3"
                    />
                  </div>
                  <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="amount" className="text-right">
                      Amount
                    </Label>
                    <Input id="amount" value={5000} className="col-span-3" />
                  </div>
                  <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="comment" className="text-right">
                      Comment
                    </Label>
                    <Input id="comment" value="" className="col-span-3" />
                  </div>
                </div>
                <DialogFooter>
                  <LoadingButton
                    type="submit"
                    onClick={() => handleSubmit()}
                    disabled={!!isSubmitting}
                    loading={isSubmitting}
                  >
                    Create
                  </LoadingButton>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </>
        }
      />
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        {recipients.map((recipient) => (
          <Card key={recipient.lightningAddress}>
            <CardHeader>
              <div className="flex flex-row items-center gap-3">
                <img src={recipient.logo} className="rounded-lg w-10 h-10" />
                <div className="flex flex-row gap-3 grow justify-between items-center">
                  <CardTitle>{recipient.name}</CardTitle>
                  <LoadingButton
                    size="sm"
                    onClick={() => {
                      setCustomName(recipient.name);
                      setCustomLightningAddress(recipient.lightningAddress);
                      setOpen(true);
                    }}
                  >
                    Support
                  </LoadingButton>
                </div>
              </div>
              <CardDescription>{recipient.description}</CardDescription>
            </CardHeader>
          </Card>
        ))}
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
