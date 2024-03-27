import React from "react";
import { Link, useNavigate } from "react-router-dom";

import Loading from "src/components/Loading";

import { request } from "src/utils/request";
import { handleRequestError } from "src/utils/handleRequestError";
import { useCSRF } from "src/hooks/useCSRF";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";
import AppHeader from "src/components/AppHeader";
import BreezRedeem from "src/components/BreezRedeem";
import { Button } from "src/components/ui/button";

function Wallet() {
  const { data: apps, mutate: mutateApps } = useApps();
  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();
  const navigate = useNavigate();
  const [showBackupPrompt, setShowBackupPrompt] = React.useState(true);

  if (!apps || !info) {
    return <Loading />;
  }

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
      handleRequestError("Failed to skip backup", error);
    } finally {
      setShowBackupPrompt(false);
    }
  }

  return (
    <>
      <AppHeader title="Wallet" description="Send and receive transactions" />

      <BreezRedeem />

      {info?.showBackupReminder && showBackupPrompt && (
        <div className="rounded-2xl bg-orange-50 border border-orange-200 flex flex-col justify-between">
          <div className="p-4 h-full border-b border-orange-200">
            <h2 className="font-medium text-orange-700 mb-2">
              Back up your recovery phrase!
            </h2>
            <p className="text-sm text-orange-700">
              Not backing up your key might result in loosing access to your
              funds.
            </p>
          </div>
          <div className="py-3 px-4 flex items-center gap-4">
            <div
              onClick={onSkipBackup}
              className="text-center font-medium p-2.5 w-full text-sm rounded-lg text-orange-700 hover:bg-orange-100 cursor-pointer"
            >
              Skip For Now
            </div>
            <div
              onClick={() => navigate("/backup/mnemonic")}
              className="text-center font-medium p-2.5 w-full text-sm rounded-lg text-orange-700 bg-orange-200 cursor-pointer hover:bg-orange-300"
            >
              Back Up Now
            </div>
          </div>
        </div>
      )}

      <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm">
        <div className="flex flex-col items-center gap-1 text-center">
          <h3 className="text-2xl font-bold tracking-tight">
            You have no funds, yet.
          </h3>
          <p className="text-sm text-muted-foreground">
            Topup your wallet and make your first transaction.
          </p>
          <Link to="/start">
            <Button className="mt-4">Get Started</Button>
          </Link>
        </div>
      </div>
    </>
  );
}

export default Wallet;
