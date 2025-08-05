import {
  CheckCircle2Icon,
  CircleXIcon,
  EditIcon,
  ExternalLinkIcon,
  InfoIcon,
  Link2Icon,
  User2Icon,
  ZapIcon,
} from "lucide-react";
import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";

import BudgetAmountSelect from "src/components/BudgetAmountSelect";
import BudgetRenewalSelect from "src/components/BudgetRenewalSelect";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import UserAvatar from "src/components/UserAvatar";
import { AppCardConnectionInfo } from "src/components/connections/AppCardConnectionInfo";
import { AppCardNotice } from "src/components/connections/AppCardNotice";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTrigger,
} from "src/components/ui/dialog";
import { LoadingButton } from "src/components/ui/loading-button";
import { Separator } from "src/components/ui/separator";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { LinkStatus, useLinkAccount } from "src/hooks/useLinkAccount";
import { BudgetRenewalType } from "src/types";

import AlbyAccountDarkSVG from "public/images/illustrations/alby-account-dark.svg";
import AlbyAccountLightSVG from "public/images/illustrations/alby-account-light.svg";
import { ALBY_ACCOUNT_APP_NAME } from "src/constants";
import { useApps } from "src/hooks/useApps";

function AlbyConnectionCard() {
  const { data: linkedAlbyAccountAppsData, mutate: reloadAlbyAccountApp } =
    useApps(undefined, undefined, {
      name: ALBY_ACCOUNT_APP_NAME,
    });
  const albyAccountApp = linkedAlbyAccountAppsData?.apps[0];
  const { data: albyMe } = useAlbyMe();
  const { loading, linkStatus, loadingLinkStatus, linkAccount } =
    useLinkAccount(reloadAlbyAccountApp);

  const [maxAmount, setMaxAmount] = useState(150_000);
  const [budgetRenewal, setBudgetRenewal] =
    useState<BudgetRenewalType>("weekly");

  function onSubmit(e: FormEvent) {
    e.preventDefault();

    linkAccount(maxAmount, budgetRenewal);
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="relative">
          Linked Alby Account
          {albyAccountApp && <AppCardNotice app={albyAccountApp} />}
        </CardTitle>
        <CardDescription>
          Link Your Alby Account to use your lightning address with Alby Hub and
          use apps that you connected to your Alby Account.
        </CardDescription>
      </CardHeader>
      <Separator />
      <CardContent className="group">
        <div className="grid grid-cols-1 xl:grid-cols-2 mt-5 gap-3 items-center relative">
          <div className="flex flex-col gap-4">
            <div className="flex flex-row gap-4">
              <UserAvatar className="h-14 w-14" />
              <div className="flex flex-col justify-center">
                <div className="text-xl font-semibold">
                  {albyMe?.name || albyMe?.email}
                </div>
                <div className="flex flex-row items-center gap-1 text-sm text-muted-foreground">
                  <ZapIcon className="size-4" />
                  {albyMe?.lightning_address}
                </div>
              </div>
            </div>
            <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
              {loadingLinkStatus && <Loading />}
              {!albyAccountApp ||
              linkStatus === LinkStatus.SharedNode ||
              linkStatus === LinkStatus.Unlinked ? (
                <Dialog>
                  <DialogTrigger asChild>
                    <LoadingButton loading={loading}>
                      {!loading && <Link2Icon />}
                      Link your Alby Account
                    </LoadingButton>
                  </DialogTrigger>
                  <DialogContent>
                    <form onSubmit={onSubmit}>
                      <DialogHeader>Link to Alby Account</DialogHeader>
                      <DialogDescription className="flex flex-col gap-4">
                        After you link your account, your lightning address and
                        every app you access through your Alby Account will
                        handle payments via the Hub.
                        <img
                          src={AlbyAccountDarkSVG}
                          className="w-full hidden dark:block"
                        />
                        <img
                          src={AlbyAccountLightSVG}
                          className="w-full dark:hidden"
                        />
                        You can add a budget that will restrict how much can be
                        spent from the Hub with your Alby Account.
                      </DialogDescription>
                      <div className="mt-4">
                        <BudgetRenewalSelect
                          value={budgetRenewal}
                          onChange={setBudgetRenewal}
                        />
                        <BudgetAmountSelect
                          value={maxAmount}
                          onChange={setMaxAmount}
                          minAmount={
                            25000 /* the minimum should be a bit more than the Alby monthly fee */
                          }
                          budgetOptions={{
                            "150k": 150_000,
                            "500k": 500_000,
                            "1M": 1_000_000,
                          }}
                        />
                      </div>
                      <DialogFooter>
                        <LoadingButton loading={loading}>
                          Link to Alby Account
                        </LoadingButton>
                      </DialogFooter>
                    </form>
                  </DialogContent>
                </Dialog>
              ) : linkStatus === LinkStatus.ThisNode ? (
                <Button
                  variant="positive"
                  disabled
                  className="disabled:opacity-100"
                >
                  <CheckCircle2Icon />
                  Alby Account Linked
                </Button>
              ) : (
                linkStatus === LinkStatus.OtherNode && (
                  <Button variant="destructive" disabled>
                    <CircleXIcon />
                    Linked to another wallet
                  </Button>
                )
              )}
              {!albyAccountApp && (
                <ExternalLink
                  to="https://www.getalby.com/node"
                  className="w-full sm:w-auto"
                >
                  <Button variant="outline" className="w-full sm:w-auto">
                    <ExternalLinkIcon />
                    Alby Account Settings
                  </Button>
                </ExternalLink>
              )}
              {albyAccountApp && (
                <Link to="/settings/alby-account" className="w-full sm:w-auto">
                  <Button variant="outline" className="w-full sm:w-auto">
                    <User2Icon />
                    Alby Account Settings
                  </Button>
                </Link>
              )}
            </div>
          </div>
          {albyAccountApp && (
            <div className="slashed-zero">
              <Link
                to={`/apps/${albyAccountApp.appPubkey}?edit=true`}
                className="absolute top-0 right-0"
              >
                <EditIcon className="size-4 hidden group-hover:inline text-muted-foreground hover:text-card-foreground" />
              </Link>
              <AppCardConnectionInfo
                connection={albyAccountApp}
                budgetRemainingText={
                  <span className="flex items-center gap-2 justify-end">
                    Left in Alby Account budget
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger>
                          <div className="flex flex-row items-center">
                            <InfoIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent className="max-w-sm">
                          Control what access your Alby Account has to your Hub
                          by editing the budget. Every app you access through
                          your Alby Account (such as your lightning address,
                          Alby Extension, podcasting 2.0 apps) will handle
                          payments via your Hub.
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </span>
                }
              />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default AlbyConnectionCard;
