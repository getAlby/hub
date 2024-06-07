import React from "react";
import { useNavigate, useParams } from "react-router-dom";

import { useApp } from "src/hooks/useApp";
import { useCSRF } from "src/hooks/useCSRF";
import { useDeleteApp } from "src/hooks/useDeleteApp";
import { useInfo } from "src/hooks/useInfo";
import { AppPermissions, BudgetRenewalType, PermissionType } from "src/types";

import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear

import AppAvatar from "src/components/AppAvatar";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import Permissions from "src/components/Permissions";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import { useToast } from "src/components/ui/use-toast";

function ShowApp() {
  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const { pubkey } = useParams() as { pubkey: string };
  const { data: app, mutate: refetchApp, error } = useApp(pubkey);
  const navigate = useNavigate();

  const [editMode, setEditMode] = React.useState(false);

  const { deleteApp, isDeleting } = useDeleteApp(() => {
    navigate("/apps");
  });

  const [permissions, setPermissions] = React.useState<AppPermissions>({
    requestMethods: new Set<PermissionType>(),
    maxAmount: 0,
    budgetRenewal: "",
    expiresAt: undefined,
  });

  React.useEffect(() => {
    if (app) {
      setPermissions({
        requestMethods: new Set(app.requestMethods as PermissionType[]),
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
      toast({ title: "Successfully updated permissions" });
    } catch (error) {
      handleRequestError(toast, "Failed to update permissions", error);
    }
  };

  if (!app) {
    return <Loading />;
  }

  return (
    <>
      <div className="w-full">
        <div className="flex flex-col gap-5">
          <AppHeader
            title={
              <div className="flex flex-row items-center">
                <AppAvatar appName={app.name} className="w-10 h-10 mr-2" />
                <h2
                  title={app.name}
                  className="text-xl font-semibold overflow-hidden text-ellipsis whitespace-nowrap"
                >
                  {app.name}
                </h2>
              </div>
            }
            contentRight={
              <AlertDialog>
                <AlertDialogTrigger>
                  <Button variant="destructive">Delete</Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This will revoke the permission and will no longer allow
                      calls from this public key.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      onClick={() => deleteApp(app.nostrPubkey)}
                      disabled={isDeleting}
                    >
                      Continue
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            }
            description={""}
          />
          <Card>
            <CardHeader>
              <CardTitle>Info</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell className="font-medium">Public Key</TableCell>
                    <TableCell className="text-muted-foreground break-all">
                      {app.nostrPubkey}
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium">Last used</TableCell>
                    <TableCell className="text-muted-foreground">
                      {app.lastEventAt
                        ? new Date(app.lastEventAt).toString()
                        : "Never"}
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium">Expires At</TableCell>
                    <TableCell className="text-muted-foreground">
                      {app.expiresAt &&
                      new Date(app.expiresAt).getFullYear() !== 1
                        ? new Date(app.expiresAt).toString()
                        : "Never"}
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
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
                                app.requestMethods as PermissionType[]
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
                canEditPermissions={editMode}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}

export default ShowApp;
