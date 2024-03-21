import React from "react";
import { Link, useNavigate } from "react-router-dom";

import Loading from "src/components/Loading";
import SuggestedApps from "src/components/SuggestedApps";
import AppCard from "src/components/AppCard";
import BreezRedeem from "src/components/BreezRedeem";
import { request } from "src/utils/request";
import { handleRequestError } from "src/utils/handleRequestError";
import { useCSRF } from "src/hooks/useCSRF";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";

function AppsList() {
  const { data: apps, mutate: mutateApps } = useApps();
  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();
  const navigate = useNavigate();
  const [showBackupPrompt, setShowBackupPrompt] = React.useState(true);

  if (!apps || !info) {
    return <Loading />;
  }

  const handleDeleteApp = (nostrPubkey: string) => {
    const updatedApps = apps.filter((app) => app.nostrPubkey !== nostrPubkey);
    mutateApps(updatedApps, false);
  };

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
    <div className="container max-w-screen-lg mt-6">
      <BreezRedeem />
      <div className="flex flex-row-reverse">
        <Link
          className="flex-row w-48 mb-6 px-0 py-2 bg-primary-gradient border-2 border-transparent text-black hover:bg-primary-gradient-hover cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary transition duration-150"
          to="/apps/new"
        >
          Add a connection
        </Link>
      </div>
      <div className="grid sm:grid-cols-2 md:grid-cols-3 gap-4">
        {!apps.length && (
          <div className="rounded-2xl bg-indigo-50 border border-indigo-200 flex flex-col justify-between">
            <div className="p-4 h-full border-b border-indigo-200">
              <h2 className="font-medium text-indigo-600 mb-2">
                Create a New Connection
              </h2>
              <p className="text-sm text-indigo-600">
                Create a new connection to connect to an NWC-powered app
              </p>
            </div>
            <div className="py-3 px-4 flex items-center gap-4">
              <Link
                to="/apps/new"
                className="text-center font-medium p-2.5 w-full text-sm rounded-lg text-indigo-600 bg-indigo-200 cursor-pointer hover:bg-indigo-300"
              >
                Create a New Connection
              </Link>
            </div>
          </div>
        )}
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
        {apps.map((app, index) => (
          <AppCard key={index} app={app} onDelete={handleDeleteApp} />
        ))}
      </div>

      <hr className="my-8" />

      <SuggestedApps />
    </div>
  );
}

export default AppsList;
