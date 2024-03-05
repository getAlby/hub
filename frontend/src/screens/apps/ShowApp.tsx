import { useNavigate, useParams } from "react-router-dom";
import gradientAvatar from "gradient-avatar";
import { PopiconsArrowLeftLine } from "@popicons/react";

import { RequestMethodType, iconMap, nip47MethodDescriptions } from "src/types";
import { useInfo } from "src/hooks/useInfo";
import { useApp } from "src/hooks/useApp";
import { useCSRF } from "src/hooks/useCSRF";
import toast from "src/components/Toast";
import Loading from "src/components/Loading";
import { request } from "src/utils/request"; // build the project for this to appear
import { handleRequestError } from "src/utils/handleRequestError";
import AppHeader from "src/components/AppHeader";
import IconButton from "src/components/IconButton";

import alby from "src/assets/suggested/alby.png";

function ShowApp() {
  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();
  const { pubkey } = useParams() as { pubkey: string };
  const { data: app, error } = useApp(pubkey);
  const navigate = useNavigate();

  if (error) {
    return <p className="text-red-500">{error.message}</p>;
  }

  if (!app || !info) {
    return <Loading />;
  }

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
      navigate("/apps");
      toast.success("App disconnected");
    } catch (error) {
      await handleRequestError("Failed to delete app", error);
    }
  };

  return (
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
          {app &&
            (app.name.toLowerCase().includes("alby") ? (
              <img
                src={alby}
                alt={app.name}
                className="block min-w-9 w-9 h-9 rounded-lg"
              />
            ) : (
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
            ))}
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
                    ? new Date(app.lastEventAt).toLocaleDateString()
                    : "never"}
                </td>
              </tr>
              <tr>
                <td className="align-top font-medium dark:text-white">
                  Expires At
                </td>
                <td className="text-gray-600 dark:text-neutral-400">
                  {app.expiresAt
                    ? new Date(app.expiresAt).toLocaleDateString()
                    : "never"}
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
          <ul className=" text-gray-600 dark:text-neutral-400">
            {app.requestMethods.map((method: string, index: number) => {
              const RequestMethodIcon = iconMap[method as RequestMethodType];
              return (
                <li key={index} className="mb-3 relative">
                  <div className="flex items-center">
                    {RequestMethodIcon && (
                      <RequestMethodIcon className="text-gray-800 dark:text-gray-300 w-4 mr-3" />
                    )}
                    <label
                      htmlFor={method}
                      className="text-gray-800 font-medium dark:text-gray-300"
                    >
                      {nip47MethodDescriptions[method as RequestMethodType]}
                    </label>
                  </div>
                </li>
              );
            })}
          </ul>
          {app.maxAmount > 0 && (
            <div className="ml-2 pl-5 border-l-2">
              <table className="text-gray-600 dark:text-neutral-400">
                <tbody>
                  <tr className="text-sm">
                    <td className="pr-2">Budget Allowance:</td>
                    <td>
                      {app.maxAmount} sats ({app.budgetUsage} sats used)
                    </td>
                  </tr>
                  <tr className="text-sm">
                    <td className="pr-2">Renews:</td>
                    <td>{app.budgetRenewal}</td>
                  </tr>
                </tbody>
              </table>
            </div>
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
                  className="flex-row w-full px-0 py-2 bg-white text-gray-700 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary transition duration-150"
                  onClick={handleDelete}
                >
                  Disconnect App
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ShowApp;
