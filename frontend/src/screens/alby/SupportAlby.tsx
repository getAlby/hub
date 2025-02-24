import { Code, PlusCircle, RefreshCw } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Button, LinkButton } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import {
  SUPPORT_ALBY_CONNECTION_NAME,
  SUPPORT_ALBY_LIGHTNING_ADDRESS,
} from "src/constants";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest, UpdateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

function SupportAlby() {
  const { data: apps } = useApps();
  const navigate = useNavigate();
  const { toast } = useToast();

  const [amount, setAmount] = React.useState("");
  const [senderName, setSenderName] = React.useState("");
  const [isSubmitting, setSubmitting] = React.useState(false);
  const [open, setOpen] = React.useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setSubmitting(true);
    try {
      if (
        apps?.some(
          (existingApp) => existingApp.name === SUPPORT_ALBY_CONNECTION_NAME
        )
      ) {
        throw new Error("A connection with the same name already exists.");
      }

      const parsedAmount = parseInt(amount);
      if (isNaN(parsedAmount) || parsedAmount < 1) {
        throw new Error("Invalid amount");
      }

      if (+amount < 1000) {
        toast({
          title: "Amount too low",
          description: "Minimum payment is 1000 sats",
          variant: "destructive",
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
        title: "Thank you for becoming a supporter",
        description: "The first payment is scheduled immediately.",
      });

      navigate("/");
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <div className="h-full w-full max-w-screen-sm mx-auto flex flex-col justify-center">
        <div className="flex flex-col items-center justify-center gap-6">
          <section className="text-center">
            <h2 className="text-3xl font-semibold mb-4 max-sm:mt-8">
              ✨ Your Support Matters
            </h2>
            <p className="text-muted-foreground text-balance">
              We are committed to elevating the Bitcoin ecosystem by offering
              reliable, efficient, and user-friendly software solutions for
              seamless transactions. With your help, we can keep pushing
              boundaries and evolving Alby Hub into something even more
              extraordinary.
            </p>
          </section>
          <Card className="w-full">
            <CardHeader>
              <CardTitle>Why Your Contribution Is Important</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="flex flex-col gap-5">
                <li className="flex flex-col">
                  <div className="flex flex-row items-center">
                    <PlusCircle className="w-4 h-4 mr-2" />
                    Unlock New Features
                  </div>
                  <div className="text-muted-foreground text-sm">
                    Your support empowers us to design and implement
                    cutting-edge{" "}
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
                    <RefreshCw className="w-4 h-4 mr-2" />
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
                    and ongoing maintenance, ensuring everything runs smoothly
                    and efficiently for all users.
                  </div>
                </li>
                <li className="flex flex-col ">
                  <div className="flex flex-row items-center">
                    <Code className="w-4 h-4 mr-2" />
                    Support Open-Source Freedom
                  </div>
                  <div className="text-muted-foreground text-sm">
                    Your support helps us keep Alby Hub true to the principles
                    of{" "}
                    <ExternalLink
                      className="underline"
                      to="https://github.com/getAlby/hub/blob/master/LICENSE"
                    >
                      free and open-source software
                    </ExternalLink>{" "}
                    and remains accessible for everyone to use, modify and
                    improve.
                  </div>
                </li>
              </ul>
            </CardContent>
          </Card>
          <AlertDialog open={open} onOpenChange={setOpen}>
            <div className="flex flex-col items-center justify-center gap-2">
              <AlertDialogTrigger asChild>
                <Button size="lg">Become a Supporter</Button>
              </AlertDialogTrigger>
              <LinkButton
                size="sm"
                variant="link"
                to="/"
                className="text-muted-foreground"
              >
                Maybe later
              </LinkButton>
            </div>
            <AlertDialogContent>
              <form onSubmit={handleSubmit}>
                <AlertDialogHeader>
                  <AlertDialogTitle>Become a Supporter</AlertDialogTitle>
                  <AlertDialogDescription>
                    A new app connection will be established to facilitate
                    monthly payments to Alby. You can cancel it anytime through
                    the connections page.
                  </AlertDialogDescription>
                </AlertDialogHeader>

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
                    <Input
                      id="sender-name"
                      value={senderName}
                      onChange={(e) => setSenderName(e.target.value)}
                      placeholder={`Nickname, npub, @twitter, etc.`}
                      className="col-span-3"
                    />
                  </div>
                </div>

                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <LoadingButton
                    type="submit"
                    disabled={!!isSubmitting}
                    loading={isSubmitting}
                  >
                    Complete Setup
                  </LoadingButton>
                </AlertDialogFooter>
              </form>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>
    </>
  );
}

export default SupportAlby;
