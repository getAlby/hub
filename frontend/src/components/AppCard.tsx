import React from "react";
import { useNavigate } from "react-router-dom";
import gradientAvatar from "gradient-avatar";
import {
  PopiconsBinLine,
  PopiconsTriangleExclamationLine,
} from "@popicons/react";

import Progressbar from "src/components/ProgressBar";
import { App, NIP_47_PAY_INVOICE_METHOD } from "src/types";
import { request } from "src/utils/request";
import toast from "src/components/Toast";
import { handleRequestError } from "src/utils/handleRequestError";

type Props = {
  app: App;
  csrf?: string;
  onDelete: (nostrPubkey: string) => void;
};

export default function AppCard({ app, csrf, onDelete }: Props) {
  const navigate = useNavigate();
  const [showPopup, setShowPopup] = React.useState(false); // State to control popup visibility

  const handleDelete = async () => {
    try {
      if (!csrf) {
        throw new Error("No CSRF token");
      }
      await request(`/api/apps/${app.nostrPubkey}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
      });
      onDelete(app.nostrPubkey);
      toast.success("App disconnected");
    } catch (error) {
      await handleRequestError("Failed to delete app", error);
    } finally {
      setShowPopup(false);
    }
  };

  const Popup = () => (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-10 flex justify-center items-center">
      <div className="rounded-xl mx-2 w-full max-w-lg bg-white border flex flex-col justify-between">
        <div className="p-4 h-full border-b flex items-center">
          <PopiconsTriangleExclamationLine className="w-20 h-20 text-red-300" />
          <div className="ml-4">
            <h2 className="font-medium text-gray-800 mb-2">
              Disconnecting <span className="font-bold">{app.name}</span>
            </h2>
            <p className="text-sm text-gray-500">
              This will revoke the permission and will no longer allow calls
              from this public key.
            </p>
          </div>
        </div>
        <div className="py-3 px-4 flex items-center gap-4">
          <button
            onClick={() => setShowPopup(false)}
            className="text-center font-medium p-2.5 w-full text-sm rounded-lg text-gray-700 hover:bg-gray-100 cursor-pointer whitespace-nowrap"
          >
            Cancel
          </button>
          <button
            onClick={handleDelete}
            className="text-center font-medium p-2.5 w-full text-sm rounded-lg text-white bg-red-500 cursor-pointer hover:bg-red-600 whitespace-nowrap"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <>
      {showPopup && <Popup />}
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
