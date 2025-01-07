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
import { CreateAppRequest, UpdateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

import { LightningAddress } from "@getalby/lightning-tools";
import { ExternalLinkIcon, PlusCircle } from "lucide-react";
import alby from "src/assets/suggested-apps/alby.png";
import hrf from "src/assets/zapplanner/hrf.png";
import opensats from "src/assets/zapplanner/opensats.png";
import ExternalLink from "src/components/ExternalLink";
import { Button, ExternalLinkButton } from "src/components/ui/button";
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
import { Textarea } from "src/components/ui/textarea";
import { request } from "src/utils/request";

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
      "Support the open-source development of Hub, Go, Lightning Browser Extension, developer tools, NWC, WebLN, Podcasting 2.0 & more.",
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
  const [isSubmitting, setSubmitting] = React.useState(false);
  const [recipientName, setRecipientName] = React.useState("");
  const [recipientLightningAddress, setRecipientLightningAddress] =
    React.useState("");
  const [amount, setAmount] = React.useState(0);
  const [comment, setComment] = React.useState("");
  const [senderName, setSenderName] = React.useState("");

  React.useEffect(() => {
    // reset form on close
    if (!open) {
      setRecipientName("");
      setRecipientLightningAddress("");
      setComment("");
      setAmount(5000);
      setSenderName("");
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      if (apps?.some((existingApp) => existingApp.name === recipientName)) {
        throw new Error("A connection with the same name already exists.");
      }

      // validate lighning address
      const ln = new LightningAddress(recipientLightningAddress);
      await ln.fetch();
      if (!ln.lnurlpData) {
        throw new Error("invalid recipient lightning address");
      }

      const maxAmount = Math.floor(amount * 1.01) + 10; // with fee reserve
      const isolated = false;

      const createAppRequest: CreateAppRequest = {
        name: `ZapPlanner - ${recipientName}`,
        scopes: ["pay_invoice"],
        budgetRenewal: "monthly",
        maxAmount,
        isolated,
        metadata: {
          app_store_app_id: "zapplanner",
          recipient_lightning_address: recipientLightningAddress,
        },
      };

      const createAppResponse = await createApp(createAppRequest);

      // TODO: proxy through hub backend and remove CSRF exceptions for zapplanner.albylabs.com
      const createSubscriptionResponse = await fetch(
        "https://zapplanner.albylabs.com/api/subscriptions",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            recipientLightningAddress: recipientLightningAddress,
            amount: amount,
            message: comment || "ZapPlanner payment from Alby Hub",
            payerData: JSON.stringify({
              ...(senderName ? { name: senderName } : {}),
            }),
            nostrWalletConnectUrl: createAppResponse.pairingUri,
            sleepDuration: "31 days",
          }),
        }
      );
      if (!createSubscriptionResponse.ok) {
        throw new Error(
          "Failed to create subscription: " + createSubscriptionResponse.status
        );
      }

      const { subscriptionId } = await createSubscriptionResponse.json();
      if (!subscriptionId) {
        throw new Error("no subscription ID in create subscription response");
      }

      // add the ZapPlanner subscription ID to the app metadata
      const updateAppRequest: UpdateAppRequest = {
        name: createAppRequest.name,
        scopes: createAppRequest.scopes,
        budgetRenewal: createAppRequest.budgetRenewal!,
        expiresAt: createAppRequest.expiresAt,
        maxAmount,
        isolated,
        metadata: {
          ...createAppRequest.metadata,
          zapplanner_subscription_id: subscriptionId,
        },
      };

      await request(`/api/apps/${createAppResponse.pairingPublicKey}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateAppRequest),
      });

      toast({
        title: "Created subscription",
        description: "The first payment is scheduled immediately.",
      });

      reloadApps();
      setOpen(false);
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    } finally {
      setSubmitting(false);
    }
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
                <form onSubmit={handleSubmit}>
                  <DialogHeader>
                    <DialogTitle>New Recurring Payment</DialogTitle>
                    <DialogDescription>
                      For advanced options go to{" "}
                      <ExternalLink
                        className="underline"
                        to="https://zapplanner.albylabs.com"
                      >
                        zapplanner.albylabs.com
                      </ExternalLink>
                    </DialogDescription>
                  </DialogHeader>

                  <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="name" className="text-right">
                        Recipient Name
                      </Label>
                      <Input
                        id="name"
                        value={recipientName}
                        required
                        onChange={(e) => setRecipientName(e.target.value)}
                        className="col-span-3"
                      />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="name" className="text-right">
                        Recipient Lightning Address
                      </Label>
                      <Input
                        id="receiver"
                        required
                        value={recipientLightningAddress}
                        onChange={(e) =>
                          setRecipientLightningAddress(e.target.value)
                        }
                        className="col-span-3"
                      />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="amount" className="text-right">
                        Amount / month (sats)
                      </Label>
                      <Input
                        id="amount"
                        value={amount}
                        onChange={(e) => setAmount(parseInt(e.target.value))}
                        className="col-span-3"
                      />
                    </div>
                    <div className="grid grid-cols-4 gap-4">
                      <Label htmlFor="comment" className="text-right pt-2">
                        Comment
                      </Label>
                      <Textarea
                        id="comment"
                        value={comment}
                        onChange={(e) => setComment(e.target.value)}
                        placeholder="Optional comment"
                        className="col-span-3"
                      />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="comment" className="text-right">
                        Your Name
                      </Label>
                      <Input
                        id="sender-name"
                        value={senderName}
                        onChange={(e) => setSenderName(e.target.value)}
                        placeholder={`Let ${recipientName || "them"} know it was from you`}
                        className="col-span-3"
                      />
                    </div>
                  </div>
                  <DialogFooter>
                    <LoadingButton
                      type="submit"
                      disabled={!!isSubmitting}
                      loading={isSubmitting}
                    >
                      Create
                    </LoadingButton>
                  </DialogFooter>
                </form>
              </DialogContent>
            </Dialog>
          </>
        }
      />
      <p className="text-muted-foreground">
        ZapPlanner is a tool to securely schedule recurring payments. A new
        special app connection with a strict budget is created for each
        scheduled payment. This allows you to securely setup recurring payments
        and be in full control.
      </p>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        {recipients.map((recipient) => (
          <Card key={recipient.lightningAddress}>
            <CardHeader>
              <div className="flex flex-row items-center gap-3">
                <img src={recipient.logo} className="rounded-lg w-10 h-10" />
                <div className="flex flex-row gap-3 grow justify-between items-center">
                  <CardTitle>{recipient.name}</CardTitle>
                  <Button
                    size="sm"
                    onClick={() => {
                      setRecipientName(recipient.name);
                      setRecipientLightningAddress(recipient.lightningAddress);
                      setOpen(true);
                    }}
                  >
                    Support
                  </Button>
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
              <AppCard
                key={index}
                app={app}
                actions={
                  app.metadata?.zapplanner_subscription_id ? (
                    <ExternalLinkButton
                      to={`https://zapplanner.albylabs.com/subscriptions/${app.metadata.zapplanner_subscription_id}`}
                      size="sm"
                    >
                      View <ExternalLinkIcon className="w-4 h-4 ml-2" />
                    </ExternalLinkButton>
                  ) : undefined
                }
              />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
