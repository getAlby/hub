import React from "react";
import { Link } from "react-router-dom";

import Loading from "src/components/Loading";

import {
  ArrowDownRightSquare,
  ArrowUpRightSquare,
  ShieldCheckIcon,
  Sparkles,
} from "lucide-react";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { useBalances } from "src/hooks/useBalances";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

function Wallet() {
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
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
      {isWalletUsable}
      <AppHeader
        title="Wallet"
        description="Send and receive transactions"
        contentRight={
          isWalletUsable && (
            <div className="text-sm text-muted-foreground flex flex-row gap-5 items-center">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <div className="flex flex-row gap-1 items-center">
                      <ArrowDownRightSquare className="w-4 h-4 " />
                      {new Intl.NumberFormat().format(
                        Math.floor(balances.lightning.totalReceivable / 1000)
                      )}{" "}
                      sats
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>Incoming liquidity</TooltipContent>
                </Tooltip>
              </TooltipProvider>
              <div className="flex flex-row gap-1 items-center">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <div className="flex flex-row gap-1 items-center">
                        <ArrowUpRightSquare className="w-4 h-4 " />
                        {new Intl.NumberFormat().format(
                          Math.floor(balances.lightning.totalSpendable / 1000)
                        )}{" "}
                        sats
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>Outgoing liquidity</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            </div>
          )
        }
      />

      <BreezRedeem />

      {info?.showBackupReminder && showBackupPrompt && (
        <>
          <Alert>
            <ShieldCheckIcon className="h-4 w-4" />
            <AlertTitle>Back up your recovery phrase!</AlertTitle>
            <AlertDescription>
              Not backing up your key might result in permanently losing access
              to your funds.
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
        <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm">
          <div className="flex flex-col items-center gap-1 text-center">
            <h3 className="text-2xl font-bold tracking-tight">
              You have no funds, yet.
            </h3>
            <p className="text-sm text-muted-foreground">
              Topup your wallet and make your first transaction.
            </p>
            <Link to="/channels/first">
              <Button className="mt-4">Get Started</Button>
            </Link>
          </div>
        </div>
      )}

      {isWalletUsable && (
        <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm">
          <div className="flex flex-col items-center gap-1 text-center">
            <Sparkles className="h-20 w-20" />
            <h3 className="text-2xl font-bold tracking-tight">
              You are ready to get started
            </h3>
            <p className="text-sm text-muted-foreground">
              Discover the ecosystem of apps
            </p>
            <Link to="/appstore">
              <Button className="mt-4">Get Started</Button>
            </Link>
          </div>
        </div>
      )}
    </>
  );
}

export default Wallet;
