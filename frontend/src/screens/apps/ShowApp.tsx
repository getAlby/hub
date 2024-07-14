import React from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";

import { useApp } from "src/hooks/useApp";
import { useCSRF } from "src/hooks/useCSRF";
import { useDeleteApp } from "src/hooks/useDeleteApp";
import {
  App,
  AppPermissions,
  BudgetRenewalType,
  UpdateAppRequest,
  WalletCapabilities,
} from "src/types";

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
import { useCapabilities } from "src/hooks/useCapabilities";
import { formatAmount } from "src/lib/utils";

function ShowApp() {
  const { pubkey } = useParams() as { pubkey: string };
  const { data: app, mutate: refetchApp, error } = useApp(pubkey);
  const { data: capabilities } = useCapabilities();

  if (error) {
    return <p className="text-red-500">{error.message}</p>;
  }

  if (!app || !capabilities) {
    return <Loading />;
  }

  return (
    <AppInternal
      app={app}
      refetchApp={refetchApp}
      capabilities={capabilities}
    />
  );
}

type AppInternalProps = {
  app: App;
  capabilities: WalletCapabilities;
  refetchApp: () => void;
};

function AppInternal({ app, refetchApp, capabilities }: AppInternalProps) {
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const navigate = useNavigate();
  const location = useLocation();
  const [editMode, setEditMode] = React.useState(false);

  React.useEffect(() => {
    const queryParams = new URLSearchParams(location.search);
    setEditMode(queryParams.has("edit"));
  }, [location.search]);

  const { deleteApp, isDeleting } = useDeleteApp(() => {
    navigate("/apps");
  });

  const [permissions, setPermissions] = React.useState<AppPermissions>({
    scopes: app.scopes,
    maxAmount: app.maxAmount,
    budgetRenewal: app.budgetRenewal as BudgetRenewalType,
    expiresAt: app.expiresAt ? new Date(app.expiresAt) : undefined,
    isolated: app.isolated,
  });

  const handleSave = async () => {
    try {
      if (!csrf) {
        throw new Error("No CSRF token");
      }

      const updateAppRequest: UpdateAppRequest = {
        scopes: Array.from(permissions.scopes),
        budgetRenewal: permissions.budgetRenewal,
        expiresAt: permissions.expiresAt?.toISOString(),
        maxAmount: permissions.maxAmount,
      };

      await request(`/api/apps/${app.nostrPubkey}`, {
        method: "PATCH",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateAppRequest),
      });

      await refetchApp();
      setEditMode(false);
      toast({ title: "Successfully updated permissions" });
    } catch (error) {
      handleRequestError(toast, "Failed to update permissions", error);
    }
  };

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
                <AlertDialogTrigger asChild>
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
                  {app.isolated && (
                    <TableRow>
                      <TableCell className="font-medium">Balance</TableCell>
                      <TableCell className="text-muted-foreground break-all">
                        {formatAmount(app.balance)} sats
                      </TableCell>
                    </TableRow>
                  )}
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
                      {app.expiresAt
                        ? new Date(app.expiresAt).toString()
                        : "Never"}
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {!app.isolated && (
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
                              window.location.reload();
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
                  capabilities={capabilities}
                  permissions={permissions}
                  setPermissions={setPermissions}
                  readOnly={!editMode}
                  isNewConnection={false}
                  budgetUsage={app.budgetUsage}
                />
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </>
  );
}

export default ShowApp;
