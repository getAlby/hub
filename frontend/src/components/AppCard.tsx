import React from "react";
import { useNavigate } from "react-router-dom";
import gradientAvatar from "gradient-avatar";
import { PopiconsBinLine } from "@popicons/react";

import Progressbar from "src/components/ProgressBar";
import DeleteConfirmationPopup from "src/components/DeleteConfirmationPopup";
import { App, NIP_47_PAY_INVOICE_METHOD } from "src/types";
import { useDeleteApp } from "src/hooks/useDeleteApp";

type Props = {
  app: App;
  csrf?: string;
  onDelete: (nostrPubkey: string) => void;
};

export default function AppCard({ app, onDelete }: Props) {
  const navigate = useNavigate();
  const [showPopup, setShowPopup] = React.useState(false);
  const { deleteApp } = useDeleteApp((nostrPubkey: string) => {
    setShowPopup(false);
    onDelete(nostrPubkey);
  });

  return (
    <>
      {showPopup && (
        <DeleteConfirmationPopup
          appName={app.name}
          onConfirm={() => deleteApp(app.nostrPubkey)}
          onCancel={() => setShowPopup(false)}
        />
      )}
      <div className="rounded-2xl bg-white border flex flex-col justify-between">
        <div
          onClick={() => navigate(`/apps/${app.nostrPubkey}`)}
          className="cursor-pointer p-4 rounded-t-2xl h-full border-b hover:bg-gray-50"
        >
          <div className="flex items-center mb-4">
            <div className="relative inline-block min-w-10 w-10 h-10 rounded-lg border">
              <img
                src={`data:image/svg+xml;base64,${btoa(
                  gradientAvatar(app.name)
                )}`}
                alt={app.name}
                className="block w-full h-full rounded-lg p-1"
              />
              <span className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 text-white text-2xl font-medium capitalize">
                {app.name.charAt(0)}
              </span>
            </div>
            <h2 className="font-semibold whitespace-nowrap text-ellipsis overflow-hidden ml-4">
              {app.name}
            </h2>
          </div>
          <div className="text-sm text-gray-700">
            {app.requestMethods?.includes(NIP_47_PAY_INVOICE_METHOD) ? (
              app.maxAmount > 0 ? (
                <>
                  <p className="mb-2">Budget Usage:</p>
                  <Progressbar
                    percentage={(app.budgetUsage * 100) / app.maxAmount}
                  />
                </>
              ) : (
                "No limits!"
              )
            ) : (
              "Payments disabled."
            )}
          </div>
        </div>
        <div className="p-4 flex items-center flex-row-reverse gap-4">
          {/* TODO: add edit option
        <div className="w-8 h-8 cursor-pointer hover:bg-gray-50 hover:border-gray-300 border rounded-full flex justify-center items-center">
          <PopiconsEditLine className="w-4 h-4 text-gray-600" />
        </div> */}
          <div
            className="w-8 h-8 cursor-pointer hover:bg-red-50 hover:border-red-200 border rounded-full flex justify-center items-center"
            onClick={() => setShowPopup(true)}
          >
            <PopiconsBinLine className="w-4 h-4 text-red-400" />
          </div>
        </div>
      </div>
    </>
  );
}
