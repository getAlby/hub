import {
  ArrowDownToDot,
  ArrowUpFromDot,
  CircleDot,
  CopyIcon,
  ExternalLink,
  ShieldCheckIcon,
  Sparkles,
  Unplug
} from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import AlbyHead from "src/assets/images/alby-head.svg";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "src/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { useToast } from "src/components/ui/use-toast";
import { localStorageKeys } from "src/constants";
import { useBalances } from "src/hooks/useBalances";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { copyToClipboard } from "src/lib/clipboard";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

function Wallet() {
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const [showBackupPrompt, setShowBackupPrompt] = React.useState(true);

  if (!info || !balances) {
    return <Loading />;
  }

  const isWalletUsable =
    balances.lightning.totalReceivable > 0 ||
    balances.lightning.totalSpendable > 0;

  async function onSkipBackup(e: React.FormEvent) {
    e.preventDefault();
    if (!csrf) {
      throw new Error("No CSRF token");
    }

    const currentDate = new Date();
    const twoWeeksLater = new Date(
      currentDate.setDate(currentDate.getDate() + 14)
    );

    try {
      await request("/api/backup-reminder", {
        method: "PATCH",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          nextBackupReminder: twoWeeksLater.toISOString(),
        }),
      });
    } catch (error) {
      handleRequestError(toast, "Failed to skip backup", error);
    } finally {
      setShowBackupPrompt(false);
    }
  }

  return (
    <>
      <AppHeader
        title="Wallet"
        description="Send and receive transactions"
        contentRight={
          isWalletUsable && (
            <>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="default">
                    <CircleDot className="mr-2 h-4 w-4 text-primary" />
                    Online
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-64" align="end">
                  <DropdownMenuItem>
                    <div className="flex flex-row gap-10 items-center w-full">
                      <div className="whitespace-nowrap flex flex-row items-center gap-2">
                        Node
                      </div>
                      <div className="overflow-hidden text-ellipsis">
                        {/* TODO: replace with skeleton loader */}
                        {nodeConnectionInfo?.pubkey || "Loading..."}
                      </div>
                      {nodeConnectionInfo && (
                        <CopyIcon
                          className="shrink-0 w-4 h-4"
                          onClick={() => {
                            copyToClipboard(nodeConnectionInfo.pubkey);
                            toast({ title: "Copied to clipboard." });
                          }}
                        />
                      )}
                    </div>
                  </DropdownMenuItem>
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuGroup>
                      <DropdownMenuLabel>Liquidity</DropdownMenuLabel>
                      <DropdownMenuItem>
                        <div className="flex flex-row gap-3 items-center justify-between w-full">
                          <div className="grid grid-flow-col gap-2 items-center">
                            <ArrowDownToDot className="w-4 h-4 " />
                            Incoming
                          </div>
                          <div className="text-muted-foreground">
                            {new Intl.NumberFormat().format(
                              Math.floor(
                                balances.lightning.totalReceivable / 1000
                              )
                            )}{" "}
                            sats
                          </div>
                        </div>
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <div className="flex flex-row gap-3 items-center justify-between w-full">
                          <div className="grid grid-flow-col gap-2 items-center">
                            <ArrowUpFromDot className="w-4 h-4 " />
                            Outgoing
                          </div>
                          <div className="text-muted-foreground">
                            {new Intl.NumberFormat().format(
                              Math.floor(
                                balances.lightning.totalSpendable / 1000
                              )
                            )}{" "}
                            sats
                          </div>
                        </div>
                      </DropdownMenuItem>
                    </DropdownMenuGroup>
                  </>
                </DropdownMenuContent>
              </DropdownMenu>
            </>
          )
        }
      />

      {!info?.onboardingCompleted && (
        <>
          {/* TODO: needs to be more visible that you need to act.
        (e.g. add it to the sidebar, have a global banner, etc) */}
          <Alert>
            <Unplug className="h-4 w-4" />
            <AlertTitle>
              Your Alby Hub is not connected to the Lightning network!
            </AlertTitle>
            <AlertDescription>
              Action required to send and receive lightning payments
              <div className="mt-3 flex items-center gap-3">
                {/* TODO: Find a better place to redirect to. 
                Onboarding is only correct if they have not migrated Alby funds yet. */}
                <Link
                  to="/"
                  onClick={() => {
                    localStorage.removeItem(localStorageKeys.onboardingSkipped);
                  }}
                >
                  <Button size="sm">Connect</Button>
                </Link>
              </div>
            </AlertDescription>
          </Alert>
        </>
      )}

      <div className="flex flex-col lg:flex-row justify-between items-start lg:items-end gap-5">
        <div className="text-5xl font-semibold">
          {new Intl.NumberFormat().format(
            Math.floor(balances.lightning.totalSpendable / 1000)
          )}{" "}
          sats
        </div>
      </div>

      {/* TODO: Enable those cards as we know how to handle different balances 
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        <Card>
          <CardHeader>
            <div className="flex flex-row justify-between">
              <CardTitle>Lightning</CardTitle>
              <ZapIcon className="w-5 h-5 text-[#FFD648]" />
            </div>
          </CardHeader>
          <CardContent>
            <span className="text-3xl font-semibold">
              {new Intl.NumberFormat().format(
                Math.floor(balances.lightning.totalSpendable / 1000)
              )}{" "}
              sats
            </span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <div className="flex flex-row justify-between">
              <CardTitle>Onchain</CardTitle>
              <BitcoinIcon className="w-5 h-5 text-[#F7931A]" />
            </div>
          </CardHeader>
          <CardContent>
            <span className="text-3xl font-semibold">
              {new Intl.NumberFormat().format(balances.onchain.total)} sats
            </span>
          </CardContent>
        </Card>
      </div>
      */}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <Link to={`https://www.getalby.com/dashboard`} target="_blank">
          <Card>
            <CardHeader>
              <div className="flex flex-row items-center">
                <img src={AlbyHead} className="w-12 h-12 rounded-xl p-1 border" />
                <div>
                  <CardTitle>
                    <h2 className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                      Alby Web
                    </h2>
                  </CardTitle>
                  <CardDescription className="ml-4">
                    You can use it as Progressive Web App on your mobile
                  </CardDescription>
                </div>
              </div>

            </CardHeader>
            <CardContent className="text-right">
              <Button variant="outline">
                Open Alby Web
                <ExternalLink className="w-4 h-4 ml-2" />
              </Button>
            </CardContent>
          </Card>
        </Link>
        <Link to={`https://www.getalby.com`} target="_blank">
          <Card>
            <CardHeader>
              <div className="flex flex-row items-center">
                <img src={AlbyHead} className="w-12 h-12 rounded-xl p-1 border bg-[#FFDF6F]" />
                <div>
                  <CardTitle>
                    <h2 className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                      Alby Browser Extension
                    </h2>
                  </CardTitle>
                  <CardDescription className="ml-4">
                    Best to use when using your favourite Internet browser
                  </CardDescription>
                </div>
              </div>

            </CardHeader>
            <CardContent className="text-right">
              <Button variant="outline">Install Alby Extension
                <ExternalLink className="w-4 h-4 ml-2" />
              </Button>
            </CardContent>
          </Card>
        </Link>
      </div>

      <BreezRedeem />

      {info?.onboardingCompleted &&
        info?.showBackupReminder &&
        showBackupPrompt && (
          <>
            <Alert>
              <ShieldCheckIcon className="h-4 w-4" />
              <AlertTitle>Back up your recovery phrase!</AlertTitle>
              <AlertDescription>
                Not backing up your key might result in permanently losing
                access to your funds.
                <div className="mt-3 flex items-center gap-3">
                  <Button onClick={onSkipBackup} variant="secondary" size="sm">
                    Skip For Now
                  </Button>
                  <Link to="/settings/backup">
                    <Button size="sm">Back Up Now</Button>
                  </Link>
                </div>
              </AlertDescription>
            </Alert>
          </>
        )}

      {isWalletUsable && (
        <>
          <EmptyState
            icon={<Sparkles />}
            title="You are ready to get started"
            description="Discover the ecosystem of apps."
            buttonText="Get Started"
            buttonLink="/appstore"
          />
        </>
      )}
    </>
  );
}

export default Wallet;
