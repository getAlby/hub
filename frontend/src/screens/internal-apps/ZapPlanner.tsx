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

import { fiat, LightningAddress } from "@getalby/lightning-tools";
import { ExternalLinkIcon, PlusCircleIcon } from "lucide-react";
import alby from "src/assets/suggested-apps/alby.png";
import bitcoinbrink from "src/assets/zapplanner/bitcoinbrink.png";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { Textarea } from "src/components/ui/textarea";
import { SUPPORT_ALBY_LIGHTNING_ADDRESS } from "src/constants";
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
      "Support the open-source development of Hub, Go, Lightning Browser Extension, developer tools and open protocols.",
    lightningAddress: SUPPORT_ALBY_LIGHTNING_ADDRESS,
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
    lightningAddress: "opensats@vlt.ge",
    logo: opensats,
  },
  {
    name: "Brink",
    description:
      "Brink exists to strengthen the Bitcoin protocol and network through fundamental research, development, funding, mentoring.",
    lightningAddress: "bitcoinbrink@zbd.gg",
    logo: bitcoinbrink,
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
  const [amount, setAmount] = React.useState("");
  const [comment, setComment] = React.useState("");
  const [senderName, setSenderName] = React.useState("");
  const [frequencyValue, setFrequencyValue] = React.useState("1");
  const [frequencyUnit, setFrequencyUnit] = React.useState("months");
  const [currency, setCurrency] = React.useState<string>("USD");
  const [currencies, setCurrencies] = React.useState<string[]>([]);

  const [convertedAmount, setConvertedAmount] = React.useState<string>("");
  const [satoshiAmount, setSatoshiAmount] = React.useState<number | undefined>(
    undefined
  );

  React.useEffect(() => {
    // fetch the fiat list and prepend sats/BTC
    async function fetchCurrencies() {
      try {
        const res = await fetch("https://getalby.com/api/rates");
        const data: Record<string, { name: string; priority: number }> =
          await res.json();
        const fiatCodes = Object.keys(data)
          // drop "BTC" - ZapPlanner uses SATS for the bitcoin currency
          .filter((code) => code !== "BTC")
          .sort((a, b) => {
            const priorityDiff = data[a].priority - data[b].priority;
            if (priorityDiff !== 0) {
              return priorityDiff;
            }
            return a.localeCompare(b);
          })
          .map((c) => c.toUpperCase());
        setCurrencies(["SATS", ...fiatCodes]);
      } catch (err) {
        console.error("Failed to load currencies", err);
      }
    }
    fetchCurrencies();
  }, []);

  React.useEffect(() => {
    // reset form on close
    if (!open) {
      setRecipientName("");
      setRecipientLightningAddress("");
      setComment("");
      setAmount("5");
      setSenderName("");
      setFrequencyValue("1");
      setFrequencyUnit("months");
      setCurrency("USD");
      setConvertedAmount("");
      setSatoshiAmount(undefined);
    }
  }, [open]);

  React.useEffect(() => {
    // If amount is empty, clear conversion output
    if (!amount) {
      setConvertedAmount("");
      setSatoshiAmount(undefined);
      return;
    }

    // Automatically convert between sats and USD if the amount changes
    const convertCurrency = async () => {
      try {
        // any fiat (not BTC) → sats
        if (currency !== "SATS") {
          const sats = await fiat.getSatoshiValue({
            amount: parseFloat(amount),
            currency: currency,
          });
          setSatoshiAmount(sats);
          setConvertedAmount(`~${sats.toLocaleString()} sats`);
        } else {
          // Convert satoshis to USD
          const sats = parseInt(amount, 10);
          setSatoshiAmount(sats);
          const fiatValue = await fiat.getFormattedFiatValue({
            satoshi: sats,
            currency: "USD",
            locale: "en-US",
          });
          setConvertedAmount(`~${fiatValue}`);
        }
      } catch (error) {
        console.error("Conversion error:", error);
        setConvertedAmount("--");
      }
    };

    convertCurrency();
  }, [amount, currency, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      if (!amount || parseFloat(amount) !== parseInt(amount)) {
        throw new Error("Amount must be a whole number");
      }
      if (!satoshiAmount) {
        throw new Error("Invalid amount");
      }
      // parse and validate the raw frequency
      const rawFreq = parseInt(frequencyValue, 10);
      if (isNaN(rawFreq) || rawFreq < 1) {
        throw new Error("Invalid frequency");
      }
      // validate lightning address
      const ln = new LightningAddress(recipientLightningAddress);
      await ln.fetch();
      if (!ln.lnurlpData) {
        throw new Error("invalid recipient lightning address");
      }
      // Determine how many payments in one month
      let periodsPerMonth: number;
      switch (frequencyUnit) {
        case "days":
          periodsPerMonth = 31 / rawFreq;
          break;
        case "weeks":
          periodsPerMonth = 31 / 7 / rawFreq;
          break;
        case "months":
          periodsPerMonth = 1 / rawFreq;
          break;
        default:
          throw new Error("Unsupported frequency unit");
      }
      periodsPerMonth = Math.ceil(periodsPerMonth);
      //  Compute raw monthly spend
      const rawSpend = satoshiAmount * periodsPerMonth;

      // with fee reserve of max(1% or 10 sats) + 30% to avoid nwc_budget_warning (see transactions service)
      const maxAmount = Math.ceil((rawSpend * 1.01 + 10) * 1.3);
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
      // months → days since months are not recognized
      const monthsToDays = (m: string) => parseInt(m, 10) * 31;

      const sleepDuration =
        frequencyUnit === "months"
          ? `${monthsToDays(frequencyValue)} days`
          : `${frequencyValue} ${frequencyUnit}`;

      // Build a “stable fiat” payload when needed
      const subscriptionBody: Record<string, unknown> = {
        recipientLightningAddress,
        message: comment || "ZapPlanner payment from Alby Hub",
        payerData: JSON.stringify({
          ...(senderName ? { name: senderName } : {}),
        }),
        nostrWalletConnectUrl: createAppResponse.pairingUri,
        sleepDuration,
        currency,
        amount: parseInt(amount),
      };

      // TODO: proxy through hub backend and remove CSRF exceptions for zapplanner.albylabs.com
      const createSubscriptionResponse = await fetch(
        "https://zapplanner.albylabs.com/api/subscriptions",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(subscriptionBody),
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
                  <PlusCircleIcon className="h-4 w-4 mr-2" />
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
                        className="col-span-3 w-70"
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
                        className="col-span-3 w-70"
                      />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="amount" className="text-right">
                        Amount
                      </Label>
                      <div className="col-span-3 flex items-center gap-2">
                        <div className="relative flex-1">
                          <Input
                            id="amount"
                            type="number"
                            min="0"
                            step="any"
                            inputMode="decimal"
                            value={amount}
                            onChange={(e) => setAmount(e.target.value)}
                            className="col-span-3 w-70"
                          />

                          {convertedAmount && (
                            <span className="absolute inset-y-0 right-3 flex items-center text-sm text-gray-500 pointer-events-none">
                              {convertedAmount}
                            </span>
                          )}
                        </div>

                        <Select value={currency} onValueChange={setCurrency}>
                          <SelectTrigger className="w-1/2">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {currencies.map((code) => (
                              <SelectItem key={code} value={code}>
                                {code === "BTC" ? "BTC (sats)" : code}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>

                    <div className="grid grid-cols-4 items-start gap-4">
                      <Label htmlFor="frequency" className="text-right pt-2">
                        Frequency
                      </Label>
                      <div className="col-span-3 flex flex-col gap-1 w-full max-w-[450px]">
                        <div className="flex items-center gap-2 w-full">
                          <Input
                            id="frequency"
                            type="number"
                            min="1"
                            step="1"
                            inputMode="numeric"
                            value={frequencyValue}
                            onChange={(e) => setFrequencyValue(e.target.value)}
                            className="col-span-3 w-70"
                          />
                          <Select
                            value={frequencyUnit}
                            onValueChange={setFrequencyUnit}
                          >
                            <SelectTrigger className="w-1/2">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="days">days</SelectItem>
                              <SelectItem value="weeks">weeks</SelectItem>
                              <SelectItem value="months">months</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>

                        <span className="text-muted-foreground text-sm">
                          Repeat payment every
                        </span>
                      </div>
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
                        className="col-span-3 w-70"
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
                        className="col-span-3 w-70"
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
