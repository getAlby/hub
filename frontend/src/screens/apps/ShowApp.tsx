import React from "react";
import { useNavigate, useParams } from "react-router-dom";
import gradientAvatar from "gradient-avatar";
import { PopiconsArrowLeftLine, PopiconsEditLine } from "@popicons/react";

import {
  AppPermissions,
  BudgetRenewalType,
  RequestMethodType,
} from "src/types";
import { useInfo } from "src/hooks/useInfo";
import { useApp } from "src/hooks/useApp";
import { useCSRF } from "src/hooks/useCSRF";
import { useDeleteApp } from "src/hooks/useDeleteApp";

import { request } from "src/utils/request"; // build the project for this to appear
import { handleRequestError } from "src/utils/handleRequestError";

import toast from "src/components/Toast";
import Loading from "src/components/Loading";
import AppHeader from "src/components/AppHeader";
import IconButton from "src/components/IconButton";
import DeleteConfirmationPopup from "src/components/DeleteConfirmationPopup";
import Permissions from "src/components/Permissions";

function ShowApp() {
  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();
  const { pubkey } = useParams() as { pubkey: string };
  const { data: app, mutate: refetchApp, error } = useApp(pubkey);
  const navigate = useNavigate();

  const [editMode, setEditMode] = React.useState(false);
  const [showPopup, setShowPopup] = React.useState(false);

  const { deleteApp } = useDeleteApp(() => {
    setShowPopup(false);
    navigate("/apps");
  });

  const [permissions, setPermissions] = React.useState<AppPermissions>({
    requestMethods: new Set<RequestMethodType>(),
    maxAmount: 0,
    budgetRenewal: "" as BudgetRenewalType,
    expiresAt: undefined as Date | undefined,
  });

  React.useEffect(() => {
    if (app) {
      setPermissions({
        requestMethods: new Set(app.requestMethods as RequestMethodType[]),
        maxAmount: app.maxAmount,
        budgetRenewal: app.budgetRenewal as BudgetRenewalType,
        expiresAt: app.expiresAt ? new Date(app.expiresAt) : undefined,
      });
    }
  }, [app]);

  if (error) {
    return <p className="text-red-500">{error.message}</p>;
  }

  if (!app || !info) {
    return <Loading />;
  }

  const handleSave = async () => {
    try {
      if (!csrf) {
        throw new Error("No CSRF token");
      }

      await request(`/api/apps/${app.nostrPubkey}`, {
        method: "PATCH",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          ...permissions,
          requestMethods: [...permissions.requestMethods].join(" "),
        }),
      });

      await refetchApp();
      setEditMode(false);
      toast.success("Permissions updated!");
    } catch (error) {
      handleRequestError("Failed to update permissions", error);
    }
  };

  return (
    <>
      {showPopup && (
        <DeleteConfirmationPopup
          appName={app.name}
          onConfirm={() => deleteApp(app.nostrPubkey)}
          onCancel={() => setShowPopup(false)}
        />
      )}
      <div className="w-full">
        <AppHeader
          headerLeft={
            <IconButton
              onClick={() => {
                navigate("/apps");
              }}
              icon={<PopiconsArrowLeftLine className="w-4 h-4" />}
            />
          }
        >
          <div className="flex-row max-w-48 sm:max-w-lg md:max-w-xl justify-center space-x-2 flex items-center">
            {app && (
              <div className="relative inline-block min-w-9 w-9 h-9 rounded-lg border">
                <img
                  src={`data:image/svg+xml;base64,${btoa(
                    gradientAvatar(app.name)
                  )}`}
                  alt={app.name}
                  className="block w-full h-full rounded-lg p-1"
                />
                <span className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 text-white text-xl font-medium capitalize">
                  {app.name.charAt(0)}
                </span>
              </div>
            )}
            <h2
              title={app ? app.name : "Fetching app..."}
              className="text-xl font-semibold dark:text-white overflow-hidden text-ellipsis whitespace-nowrap my-2"
            >
              {app ? app.name : "Fetching app..."}
            </h2>
          </div>
        </AppHeader>

        <div className="max-w-screen-lg mx-auto">
          <div className="flex justify-between items-center pt-8 pb-4">
            <h2 className="text-2xl font-bold font-headline dark:text-white">
              App Overview
            </h2>
          </div>

          <div className="bg-white rounded-md shadow p-4 md:p-6 dark:bg-surface-02dp">
            <table>
              <tbody>
                <tr>
                  <td className="align-top w-32 font-medium dark:text-white">
                    Public Key
                  </td>
                  <td className="text-gray-600 dark:text-neutral-400 break-all">
                    {app.nostrPubkey}
                  </td>
                </tr>
                <tr>
                  <td className="align-top font-medium dark:text-white">
                    Last used
                  </td>
                  <td className="text-gray-600 dark:text-neutral-400">
                    {app.lastEventAt
                      ? new Date(app.lastEventAt).toString()
                      : "Never"}
                  </td>
                </tr>
                <tr>
                  <td className="align-top font-medium dark:text-white">
                    Expires At
                  </td>
                  <td className="text-gray-600 dark:text-neutral-400">
                    {app.expiresAt &&
                    new Date(app.expiresAt).getFullYear() !== 1
                      ? new Date(app.expiresAt).toString()
                      : "Never"}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div className="flex justify-between items-center pt-8 pb-4">
            <h2 className="text-2xl font-bold font-headline dark:text-white">
              🔒 Permissions
            </h2>
          </div>

          <div className="bg-white rounded-md shadow p-4 md:p-6 dark:bg-surface-02dp">
            <Permissions
              initialPermissions={permissions}
              onPermissionsChange={setPermissions}
              budgetUsage={app.budgetUsage}
              isEditing={editMode}
            />

            {editMode ? (
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  className="mt-6 flex-row px-6 py-2 bg-white border border-indigo-500  text-indigo-500 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary transition duration-150"
                  onClick={() => {
                    setPermissions({
                      requestMethods: new Set(
                        app.requestMethods as RequestMethodType[]
                      ),
                      maxAmount: app.maxAmount,
                      budgetRenewal: app.budgetRenewal as BudgetRenewalType,
                      expiresAt: app.expiresAt
                        ? new Date(app.expiresAt)
                        : undefined,
                    });
                    setEditMode(!editMode);
                    // TODO: clicking cancel and then editing again will leave the days option wrong
                  }}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="mt-6 flex-row px-6 py-2 bg-indigo-500 border text-white dark:bg-indigo-700 hover:bg-indigo-600 cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary transition duration-150"
                  onClick={handleSave}
                >
                  Save
                </button>
              </div>
            ) : (
              <button
                type="button"
                className="mt-6 flex-row px-6 py-2 bg-white border border-indigo-500  text-indigo-500 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary transition duration-150"
                onClick={() => setEditMode(!editMode)}
              >
                <PopiconsEditLine className="text-indigo-500 dark:bg-surface-02dp w-4 mr-3" />
                Edit Permissions
              </button>
            )}
          </div>

          <div className="relative flex py-5 mt-8 items-center">
            <div className="flex-grow border-t border-gray-300 dark:border-gray-700"></div>
            <span className="flex-shrink mx-4 text-gray-600 dark:text-gray-400 fw-bold">
              ⛔️ Danger Zone
            </span>
            <div className="flex-grow border-t border-gray-300 dark:border-gray-700"></div>
          </div>
          <div className="shadow bg-white rounded-md sm:overflow-hidden mb-5 px-6 py-2 divide-y divide-gray-200 dark:divide-neutral-700 dark:bg-surface-01dp">
            <div className="flex-col sm:flex-row flex justify-between py-4">
              <div>
                <span className="text-black dark:text-white font-medium">
                  Remove This Connection
                </span>
                <p className="text-gray-600 mr-1 dark:text-neutral-400 text-sm">
                  This will revoke the permission and will no longer allow calls
                  from this public key.
                </p>
              </div>
              <div className="flex items-center">
                <div className="sm:w-64 flex-none w-full pt-4 sm:pt-0">
                  <button
                    type="button"
                    className="flex-row w-full px-0 py-2 bg-white border border-red-500 text-red-500 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary transition duration-150"
                    onClick={() => setShowPopup(true)}
                  >
                    Disconnect App
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

export default ShowApp;
