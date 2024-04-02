import gradientAvatar from "gradient-avatar";
import React from "react";
import { useNavigate, useParams } from "react-router-dom";

import { useApp } from "src/hooks/useApp";
import { useCSRF } from "src/hooks/useCSRF";
import { useDeleteApp } from "src/hooks/useDeleteApp";
import { useInfo } from "src/hooks/useInfo";
import {
  AppPermissions,
  BudgetRenewalType,
  RequestMethodType,
} from "src/types";

import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear

import DeleteConfirmationPopup from "src/components/DeleteConfirmationPopup";
import Loading from "src/components/Loading";
import Permissions from "src/components/Permissions";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";

function ShowApp() {
  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
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
    budgetRenewal: "",
    expiresAt: undefined,
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
      toast({ title: "Permissions updated!" });
    } catch (error) {
      handleRequestError(toast, "Failed to update permissions", error);
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
        <div className="flex flex-col gap-5">
          <div className="flex flex-row items-center justify-between">
            <div className="flex flex-row items-center ">
              {app && (
                <div className="relative inline-block min-w-9 w-9 h-9 rounded-lg border mr-2">
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
                className="text-xl font-semibold overflow-hidden text-ellipsis whitespace-nowrap my-2"
              >
                {app ? app.name : "Fetching app..."}
              </h2>
            </div>
            <Button variant="destructive" onClick={() => setShowPopup(true)}>
              Delete
            </Button>
          </div>
          <Card>
            <CardHeader>
              <CardTitle>Info</CardTitle>
            </CardHeader>
            <CardContent>
              <table>
                <tbody>
                  <tr>
                    <td className="align-top w-32 font-medium">Public Key</td>
                    <td className="text-muted-foreground break-all">
                      {app.nostrPubkey}
                    </td>
                  </tr>
                  <tr>
                    <td className="align-top font-medium">Last used</td>
                    <td className="text-gray-600 text-muted-foreground">
                      {app.lastEventAt
                        ? new Date(app.lastEventAt).toString()
                        : "Never"}
                    </td>
                  </tr>
                  <tr>
                    <td className="align-top font-medium">Expires At</td>
                    <td className="text-muted-foreground">
                      {app.expiresAt &&
                      new Date(app.expiresAt).getFullYear() !== 1
                        ? new Date(app.expiresAt).toString()
                        : "Never"}
                    </td>
                  </tr>
                </tbody>
              </table>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>
                <div className="flex flex-row justify-between items-center">
                  Permissions
                  <div className="flex flex-row gap-2">
                    {editMode && (
                      <div className="flex justify-center items-center gap-2">
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => {
                            setPermissions({
                              requestMethods: new Set(
                                app.requestMethods as RequestMethodType[]
                              ),
                              maxAmount: app.maxAmount,
                              budgetRenewal:
                                app.budgetRenewal as BudgetRenewalType,
                              expiresAt: app.expiresAt
                                ? new Date(app.expiresAt)
                                : undefined,
                            });
                            setEditMode(!editMode);
                            // TODO: clicking cancel and then editing again will leave the days option wrong
                          }}
                        >
                          Cancel
                        </Button>

                        <Button type="button" onClick={handleSave}>
                          Save
                        </Button>
                      </div>
                    )}

                    {!editMode && (
                      <>
                        <Button
                          variant="outline"
                          onClick={() => setEditMode(!editMode)}
                        >
                          Edit
                        </Button>
                      </>
                    )}
                  </div>
                </div>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Permissions
                initialPermissions={permissions}
                onPermissionsChange={setPermissions}
                budgetUsage={app.budgetUsage}
                isEditing={editMode}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}

export default ShowApp;
