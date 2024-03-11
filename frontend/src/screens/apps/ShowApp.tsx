import React from "react";
import { useNavigate, useParams } from "react-router-dom";
import gradientAvatar from "gradient-avatar";
import { PopiconsArrowLeftLine, PopiconsEditLine } from "@popicons/react";

import {
  RequestMethodType,
  budgetOptions,
  iconMap,
  nip47MethodDescriptions,
} from "src/types";
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
  const { data: app, mutate: refetchInfo, error } = useApp(pubkey);
  const navigate = useNavigate();

  const [editMode, setEditMode] = React.useState(false);
  const [maxAmount, setMaxAmount] = React.useState(0);

  const parseRequestMethods = (reqParam: string): Set<RequestMethodType> => {
    const methods = reqParam
      ? reqParam.split(" ")
      : Object.keys(nip47MethodDescriptions);
    // Create a Set of RequestMethodType from the array
    const requestMethodsSet = new Set<RequestMethodType>(
      methods as RequestMethodType[]
    );
    return requestMethodsSet;
  };

  const [requestMethods, setRequestMethods] = React.useState(
    parseRequestMethods("")
  );

  React.useEffect(() => {
    if (app) {
      setMaxAmount(app.maxAmount);
      setRequestMethods(parseRequestMethods(app.requestMethods.join(" ")));
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

      const requestMethodsToAdd = new Set<RequestMethodType>();
      requestMethods.forEach((rm) => {
        if (!app.requestMethods.includes(rm)) {
          requestMethodsToAdd.add(rm);
        }
      });

      const requestMethodsToRemove = new Set<RequestMethodType>();
      (app.requestMethods as RequestMethodType[]).forEach((rm) => {
        if (!requestMethods.has(rm)) {
          requestMethodsToRemove.add(rm);
        }
      });

      await request(`/api/apps/${app.nostrPubkey}`, {
        method: "PATCH",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          maxAmount,
          budgetRenewal: app.budgetRenewal || "monthly",
          expiresAt: app.expiresAt,
          requestMethodsToAdd: [...requestMethodsToAdd].join(" "),
          requestMethodsToRemove: [...requestMethodsToRemove].join(" "),
        }),
      });

      await refetchInfo();
      setEditMode(false);
      toast.success("Permissions updated!");
    } catch (error) {
      handleRequestError("Failed to update permissions", error);
    }
  };

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

  const isEdited = () => {
    const requestMethodCheck =
      app.requestMethods.length === requestMethods.size &&
      app.requestMethods.every((rm) =>
        requestMethods.has(rm as RequestMethodType)
      );
    const maxAmountCheck = maxAmount === app.maxAmount;
    return !(requestMethodCheck && maxAmountCheck);
  };

  const handleRequestMethodChange = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const requestMethod = event.target.value as RequestMethodType;

    const newRequestMethods = new Set(requestMethods);

    if (newRequestMethods.has(requestMethod)) {
      // If checked and item is already in the list, remove it
      newRequestMethods.delete(requestMethod);
    } else {
      newRequestMethods.add(requestMethod);
    }

    setRequestMethods(newRequestMethods);
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
          <>
            <ul className="flex flex-col w-full">
              {(
                Object.keys(nip47MethodDescriptions) as RequestMethodType[]
              ).map((rm, index) => {
                const RequestMethodIcon = iconMap[rm];
                return (
                  <li
                    key={index}
                    className={`w-full ${
                      rm == "pay_invoice" ? "order-last" : ""
                    } ${!editMode && !requestMethods.has(rm) ? "hidden" : ""}`}
                  >
                    <div className="flex items-center mb-2">
                      {RequestMethodIcon && (
                        <RequestMethodIcon
                          className={`text-gray-800 dark:text-gray-300 w-4 mr-3 ${
                            editMode ? "hidden" : ""
                          }`}
                        />
                      )}
                      <input
                        type="checkbox"
                        id={rm}
                        value={rm}
                        checked={requestMethods.has(rm as RequestMethodType)}
                        onChange={handleRequestMethodChange}
                        className={` ${
                          !editMode ? "hidden" : ""
                        } w-4 h-4 mr-4 text-indigo-500 bg-gray-50 border border-gray-300 rounded focus:ring-indigo-500 dark:focus:ring-indigo-400 dark:ring-offset-gray-800 focus:ring-2 dark:bg-surface-00dp dark:border-gray-700 cursor-pointer`}
                      />
                      <label
                        htmlFor={rm}
                        className="text-gray-800 dark:text-gray-300 cursor-pointer"
                      >
                        {nip47MethodDescriptions[rm as RequestMethodType]}
                      </label>
                    </div>
                    {rm == "pay_invoice" &&
                      (editMode ? (
                        <div
                          className={`pt-2 pb-2 pl-5 ml-2.5 border-l-2 border-l-gray-200 dark:border-l-gray-400 ${
                            !requestMethods.has(rm)
                              ? "pointer-events-none opacity-30"
                              : ""
                          }`}
                        >
                          <p className="text-gray-600 dark:text-gray-300 mb-2 text-sm">
                            Monthly budget
                          </p>
                          <div
                            id="budget-allowance-limits"
                            className="grid grid-cols-6 grid-rows-2 md:grid-rows-1 md:grid-cols-6 gap-2 text-xs text-gray-800 dark:text-neutral-200"
                          >
                            {Object.keys(budgetOptions).map((budget) => {
                              return (
                                <div
                                  key={budget}
                                  onClick={() =>
                                    setMaxAmount(budgetOptions[budget])
                                  }
                                  className={`col-span-2 md:col-span-1 cursor-pointer rounded border-2 ${
                                    maxAmount == budgetOptions[budget]
                                      ? "border-indigo-500 dark:border-indigo-400 text-indigo-500 bg-indigo-100 dark:bg-purple-900"
                                      : "border-gray-200 dark:border-gray-400"
                                  } text-center py-4 dark:text-white`}
                                >
                                  {budget}
                                  <br />
                                  {budgetOptions[budget] ? "sats" : "#reckless"}
                                </div>
                              );
                            })}
                          </div>
                        </div>
                      ) : (
                        <div className="ml-2 pl-5 border-l-2">
                          <table className="text-gray-600 dark:text-neutral-400">
                            <tbody>
                              <tr className="text-sm">
                                <td className="pr-2">Budget Allowance:</td>
                                <td>
                                  {app.maxAmount || "∞"} sats ({app.budgetUsage}{" "}
                                  sats used)
                                </td>
                              </tr>
                              <tr className="text-sm">
                                <td className="pr-2">Renews:</td>
                                <td>{app.budgetRenewal}</td>
                              </tr>
                            </tbody>
                          </table>
                        </div>
                      ))}
                  </li>
                );
              })}
            </ul>
          </>

          {editMode ? (
            <button
              type="button"
              className="mt-6 flex-row px-6 py-2 bg-white border border-indigo-500  text-indigo-500 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary transition duration-150"
              disabled={!isEdited()}
              onClick={handleSave}
            >
              Save
            </button>
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
