import {
  CodeIcon,
  HandCoins,
  PlusCircleIcon,
  RefreshCwIcon,
  Sparkles,
} from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
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
import { UpgradeDialog } from "src/components/UpgradeDialog";
import {
  BITCOIN_DISPLAY_FORMAT_BIP177,
  SUPPORT_ALBY_CONNECTION_NAME,
  SUPPORT_ALBY_LIGHTNING_ADDRESS,
} from "src/constants";
import { useInfo } from "src/hooks/useInfo";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest, UpdateAppRequest } from "src/types";
import { formatBitcoinAmount } from "src/utils/bitcoinFormatting";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

function SupportAlby() {
  const navigate = useNavigate();
  const { data: info } = useInfo();

  const [amount, setAmount] = React.useState("");
  const [senderName, setSenderName] = React.useState("");
  const [isSubmitting, setSubmitting] = React.useState(false);
  const [open, setOpen] = React.useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setSubmitting(true);
    try {
      const parsedAmount = parseInt(amount);
      if (isNaN(parsedAmount) || parsedAmount < 1) {
        throw new Error("Invalid amount");
      }

      if (+amount < 1000) {
        toast.error("Amount too low", {
          description: `Minimum payment is ${formatBitcoinAmount(
            1_000 * 1000,
            info?.bitcoinDisplayFormat || BITCOIN_DISPLAY_FORMAT_BIP177
          )}`,
        });
        return;
      }

      // TODO: extract below code as is duplicated with ZapPlanner
      // with fee reserve of max(1% or 10 sats) + 30% to avoid nwc_budget_warning (see transactions service)
      const maxAmount = Math.floor((parsedAmount * 1.01 + 10) * 1.3);
      const isolated = false;

      const createAppRequest: CreateAppRequest = {
        name: SUPPORT_ALBY_CONNECTION_NAME,
        scopes: ["pay_invoice"],
        budgetRenewal: "monthly",
        maxAmount,
        isolated,
        metadata: {
          app_store_app_id: "zapplanner",
          recipient_lightning_address: SUPPORT_ALBY_LIGHTNING_ADDRESS,
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
            recipientLightningAddress: SUPPORT_ALBY_LIGHTNING_ADDRESS,
            amount: parsedAmount,
            message: "ZapPlanner payment from Alby Hub",
            payerData: JSON.stringify({
              ...(senderName ? { name: senderName } : {}),
            }),
            nostrWalletConnectUrl: createAppResponse.pairingUri,
            cronExpression: "0 0 1 * *", // at the start of each month
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
      // Only send metadata since that's the only thing changing
      const updateAppRequest: UpdateAppRequest = {
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

      toast("Thank you for becoming a supporter", {
        description: "Payment will be made at the start of each month",
      });

      navigate("/");
    } catch (error) {
      handleRequestError("Failed to create app", error);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <AppHeader
        title="Support Alby Hub"
        description="We are committed to elevating the Bitcoin ecosystem by offering reliable, efficient, and user-friendly software solutions for seamless transactions. With your help, we can keep pushing boundaries and evolving Alby Hub into something extraordinary."
      />
      <h2 className="text-2xl font-semibold">Become a Supporter</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card className="flex flex-col">
          <CardHeader className="grow">
            <CardTitle>Upgrade to Pro</CardTitle>
            <CardDescription>
              Upgrade your Alby Account to Pro for a small fee and enjoy
              additional perks that come with it!
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-end">
            <UpgradeDialog>
              <Button>
                <Sparkles />
                Upgrade to Pro
              </Button>
            </UpgradeDialog>
          </CardFooter>
        </Card>
        <Card className="flex flex-col">
          <CardHeader className="grow">
            <CardTitle>Donate to Alby Hub development</CardTitle>
            <CardDescription>
              Set up a recurring value4value payment to support the development
              of Alby Hub, Alby Go, and the NWC ecosystem.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 grow" />
          <CardFooter className="flex justify-end">
            <Dialog open={open} onOpenChange={setOpen}>
              <div className="flex flex-col items-center justify-center gap-2">
                <DialogTrigger asChild>
                  <Button>
                    <HandCoins />
                    Setup Donation
                  </Button>
                </DialogTrigger>
              </div>
              <DialogContent>
                <form onSubmit={handleSubmit}>
                  <DialogHeader>
                    <DialogTitle>Become a Supporter</DialogTitle>
                    <DialogDescription>
                      A new app connection will be established to facilitate
                      monthly payments to Alby. You can cancel it anytime
                      through the connections page.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="flex flex-col gap-3 my-5">
                    <div className="grid grid-cols-4 gap-4">
                      <Label htmlFor="amount" className="text-right mt-2">
                        Amount <br></br>
                        <span className="font-normal text-muted-foreground">
                          (sats / month)
                        </span>
                      </Label>
                      <div className="col-span-3">
                        <Input
                          id="amount"
                          value={amount}
                          required
                          onChange={(e) => setAmount(e.target.value)}
                        />
                        <div className="grid grid-cols-3 gap-1 mt-1">
                          <Button
                            type="button"
                            variant="outline"
                            onClick={() => setAmount("3000")}
                          >
                            🙏 3000
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            onClick={() => setAmount("6000")}
                          >
                            💪 6000
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            onClick={() => setAmount("10000")}
                          >
                            ✨ 10000
                          </Button>
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="comment" className="text-right">
                        Name{" "}
                        <span className="font-normal text-muted-foreground">
                          (optional)
                        </span>
                      </Label>
                      <div className="col-span-3">
                        <Input
                          id="sender-name"
                          value={senderName}
                          onChange={(e) => setSenderName(e.target.value)}
                          placeholder={`Nickname, npub, @twitter, etc.`}
                        />
                      </div>
                    </div>
                  </div>

                  <DialogFooter>
                    <LoadingButton
                      type="submit"
                      disabled={!!isSubmitting}
                      loading={isSubmitting}
                    >
                      Complete Setup
                    </LoadingButton>
                  </DialogFooter>
                </form>
              </DialogContent>
            </Dialog>
          </CardFooter>
        </Card>
      </div>
      <div className="mt-4">
        <h2 className="text-2xl font-semibold mb-4">
          Why Your Contribution Is Important
        </h2>
        <ul className="flex flex-col gap-5">
          <li className="flex flex-col">
            <div className="flex flex-row items-center">
              <PlusCircleIcon className="size-4 mr-2" />
              Unlock New Features
            </div>
            <div className="text-muted-foreground text-sm">
              Your support empowers us to design and implement cutting-edge{" "}
              <ExternalLink
                className="underline"
                to="https://github.com/getAlby/hub/issues"
              >
                features
              </ExternalLink>{" "}
              that enhance your experience and keep us at the forefront of
              technology.
            </div>
          </li>
          <li className="flex flex-col ">
            <div className="flex flex-row items-center">
              <RefreshCwIcon className="size-4 mr-2" />
              Ensure Continuous Improvement
            </div>
            <div className="text-muted-foreground text-sm">
              With your contributions, we can provide{" "}
              <ExternalLink
                className="underline"
                to="https://github.com/getAlby/hub/releases"
              >
                regular updates
              </ExternalLink>{" "}
              and ongoing maintenance, ensuring everything runs smoothly and
              efficiently for all users.
            </div>
          </li>
          <li className="flex flex-col ">
            <div className="flex flex-row items-center">
              <CodeIcon className="size-4 mr-2" />
              Support Open-Source Freedom
            </div>
            <div className="text-muted-foreground text-sm">
              Your support helps us keep Alby Hub true to the principles of{" "}
              <ExternalLink
                className="underline"
                to="https://github.com/getAlby/hub/blob/master/LICENSE"
              >
                free and open-source software
              </ExternalLink>{" "}
              and remains accessible for everyone to use, modify and improve.
            </div>
          </li>
        </ul>
      </div>
    </>
  );
}

export default SupportAlby;
