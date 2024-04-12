import React from "react";
import { Link } from "react-router-dom";

import Loading from "src/components/Loading";

import {
  ArrowDownToDot,
  ArrowUpFromDot,
  CopyIcon,
  Dot,
  ShieldCheckIcon,
  Sparkles,
  Unplug,
  WalletIcon,
} from "lucide-react";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import EmptyState from "src/components/EmptyState";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
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
                    <Dot className="mr-2 h-4 w-4 text-primary" />
                    Connected
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-64">
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

      <BreezRedeem />

      {!info?.onboardingCompleted && (
        <>
          {/* TODO: needs to be more visible that you need to act.
        (e.g. add it to the sidebar, have a global banner, etc) */}
          <Alert>
            <Unplug className="h-4 w-4" />
            <AlertTitle>
              You are not connected to the lightning network!
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
                  <Link to="/backup/mnemonic">
                    <Button size="sm">Back Up Now</Button>
                  </Link>
                </div>
              </AlertDescription>
            </Alert>
          </>
        )}

      {!isWalletUsable && (
        <EmptyState
          icon={<WalletIcon />}
          title="You have no funds, yet"
          description="Topup your wallet and make your first transaction."
          buttonText="Get Started"
          buttonLink="/channels/first"
        />
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
