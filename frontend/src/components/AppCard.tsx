import React from "react";
import { Link } from "react-router-dom";
import gradientAvatar from "gradient-avatar";
import { PopiconsBinLine, PopiconsEditLine } from "@popicons/react";

import DeleteConfirmationPopup from "src/components/DeleteConfirmationPopup";
import { App, NIP_47_PAY_INVOICE_METHOD } from "src/types";
import { useDeleteApp } from "src/hooks/useDeleteApp";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress";

type Props = {
  app: App;
  csrf?: string;
  onDelete: (nostrPubkey: string) => void;
};

export default function AppCard({ app, onDelete }: Props) {
  const [showDeletePopup, setShowDeletePopup] = React.useState(false);
  const { deleteApp } = useDeleteApp((nostrPubkey: string) => {
    setShowDeletePopup(false);
    onDelete(nostrPubkey);
  });

  return (
    <>
      {showDeletePopup && (
        <DeleteConfirmationPopup
          appName={app.name}
          onConfirm={() => deleteApp(app.nostrPubkey)}
          onCancel={() => setShowDeletePopup(false)}
        />
      )}

      <Link to={`/connections/${app.nostrPubkey}`}>
        <Card>
          <CardHeader>
            <CardTitle>
              <div className="flex flex-row items-center">
                <div className="relative w-10 h-10 rounded-lg border">
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
                <h2 className="flex-1 font-semibold whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                  {app.name}
                </h2>
                <div className="flex gap-4">
                  <Link to={`/connections/${app.nostrPubkey}`}>
                    <div className="w-8 h-8 cursor-pointer hover:bg-gray-50 hover:border-gray-300 border rounded-full flex justify-center items-center">
                      <PopiconsEditLine className="w-4 h-4 text-gray-600" />
                    </div>
                  </Link>
                  <div
                    className="w-8 h-8 cursor-pointer hover:bg-red-50 hover:border-red-200 border rounded-full flex justify-center items-center"
                    onClick={() => setShowDeletePopup(true)}
                  >
                    <PopiconsBinLine className="w-4 h-4 text-red-400" />
                  </div>
                </div>
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {app.requestMethods?.includes(NIP_47_PAY_INVOICE_METHOD) ? (
              app.maxAmount > 0 ? (
                <>
                  <p className="mb-2">
                    You've spent:
                    <br />
                    {new Intl.NumberFormat().format(app.budgetUsage)} sats
                  </p>
                  <Progress
                    className="h-4"
                    value={(app.budgetUsage * 100) / app.maxAmount}
                  />
                </>
              ) : (
                "No limits!"
              )
            ) : (
              "Payments disabled."
            )}
          </CardContent>
        </Card>
      </Link>

      {/* <div className="rounded-2xl bg-white border flex flex-col justify-between">
        <Link to={`/apps/${app.nostrPubkey}`}>
          <div className="cursor-pointer p-4 rounded-t-2xl h-full border-b hover:bg-gray-50">
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
                    <p className="mb-2">
                      Budget Usage:{" "}
                      {new Intl.NumberFormat().format(app.budgetUsage)} sats
                    </p>
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
        </Link>
        <div className="p-4 flex items-center justify-end flex-row gap-4">
          <Link to={`/apps/${app.nostrPubkey}`}>
            <div className="w-8 h-8 cursor-pointer hover:bg-gray-50 hover:border-gray-300 border rounded-full flex justify-center items-center">
              <PopiconsEditLine className="w-4 h-4 text-gray-600" />
            </div>
          </Link>
          <div
            className="w-8 h-8 cursor-pointer hover:bg-red-50 hover:border-red-200 border rounded-full flex justify-center items-center"
            onClick={() => setShowDeletePopup(true)}
          >
            <PopiconsBinLine className="w-4 h-4 text-red-400" />
          </div>
        </div>
      </div> */}
    </>
  );
}
