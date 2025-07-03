import React from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";

import { useApp } from "src/hooks/useApp";

import { useDeleteApp } from "src/hooks/useDeleteApp";
import {
  App,
  AppPermissions,
  UpdateAppRequest,
  WalletCapabilities,
} from "src/types";

import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear

import { AlertCircleIcon, PencilIcon, Trash2Icon } from "lucide-react";
import AppAvatar from "src/components/AppAvatar";
import AppHeader from "src/components/AppHeader";
import { IsolatedAppTopupDialog } from "src/components/IsolatedAppTopupDialog";
import Loading from "src/components/Loading";
import Permissions from "src/components/Permissions";
import TransactionsList from "src/components/TransactionsList";
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
import { Input } from "src/components/ui/input";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useCapabilities } from "src/hooks/useCapabilities";

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
  const { toast } = useToast();
  const navigate = useNavigate();
  const location = useLocation();
  const [isEditingName, setIsEditingName] = React.useState(false);
  const [isEditingPermissions, setIsEditingPermissions] = React.useState(false);

  React.useEffect(() => {
    const queryParams = new URLSearchParams(location.search);
    const editMode = queryParams.has("edit");
    setIsEditingPermissions(editMode);
  }, [location.search]);

  const { deleteApp, isDeleting } = useDeleteApp(() => {
    navigate(
      app.metadata?.app_store_app_id !== SUBWALLET_APPSTORE_APP_ID
        ? "/apps"
        : "/sub-wallets"
    );
  });

  const [name, setName] = React.useState(app.name);
  const [permissions, setPermissions] = React.useState<AppPermissions>({
    scopes: app.scopes,
    maxAmount: app.maxAmount,
    budgetRenewal: app.budgetRenewal,
    expiresAt: app.expiresAt ? new Date(app.expiresAt) : undefined,
    isolated: app.isolated,
  });

  const handleSave = async () => {
    try {
      const updateAppRequest: UpdateAppRequest = {
        name,
        scopes: Array.from(permissions.scopes),
        budgetRenewal: permissions.budgetRenewal,
        expiresAt: permissions.expiresAt?.toISOString(),
        maxAmount: permissions.maxAmount,
        isolated: permissions.isolated,
      };

      await request(`/api/apps/${app.appPubkey}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateAppRequest),
      });

      await refetchApp();
      setIsEditingName(false);
      setIsEditingPermissions(false);
      toast({
        title: "Successfully updated connection",
      });
    } catch (error) {
      handleRequestError(toast, "Failed to update connection", error);
    }
  };

  const appName = app.name === "getalby.com" ? "Alby Account" : app.name;

  return (
    <>
      <div className="w-full">
        <div className="flex flex-col gap-5">
          <AppHeader
            title={
              <div className="flex flex-row items-center">
                <AppAvatar app={app} className="w-10 h-10 mr-2" />
                {isEditingName ? (
                  <div className="flex flex-row gap-2 items-center">
                    <Input
                      autoFocus
                      type="text"
                      name="name"
                      value={name}
                      id="name"
                      onChange={(e) => setName(e.target.value)}
                      required
                      className="text-xl font-semibold w-max max-w-40 md:max-w-fit"
                      autoComplete="off"
                    />
                    <Button type="button" onClick={handleSave}>
                      Save
                    </Button>
                  </div>
                ) : (
                  <div
                    className="flex flex-row gap-2 items-center cursor-pointer"
                    onClick={() => setIsEditingName(true)}
                  >
                    <h2
                      title={appName}
                      className="text-xl font-semibold overflow-hidden text-ellipsis whitespace-nowrap"
                    >
                      {appName}
                    </h2>
                    {app.name !== "getalby.com" && (
                      <PencilIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                    )}
                  </div>
                )}
              </div>
            }
            contentRight={
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive" size="icon">
                    <Trash2Icon className="w-4 h-4" />
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>
                      Are you sure you want to delete this connection?
                    </AlertDialogTitle>
                    <AlertDialogDescription>
                      Connected apps will no longer be able to use this
                      connection.
                      {app.isolated && (
                        <>
                          {" "}
                          No funds will be lost during this process, the balance
                          will remain in your wallet.
                        </>
                      )}
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      onClick={() => deleteApp(app.appPubkey)}
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
              <CardTitle>Connection Summary</CardTitle>
            </CardHeader>
            <CardContent className="slashed-zero">
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell className="font-medium">Id</TableCell>
                    <TableCell className="text-muted-foreground break-all">
                      {app.id}
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium">
                      App Public Key
                    </TableCell>
                    <TableCell className="text-muted-foreground break-all">
                      {app.appPubkey}
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium">
                      Wallet Public Key
                    </TableCell>
                    <TableCell className="text-muted-foreground flex items-center">
                      <span className="break-all">{app.walletPubkey}</span>
                      {!app.uniqueWalletPubkey && (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <AlertCircleIcon className="w-3 h-3 ml-2 flex-shrink-0" />
                            </TooltipTrigger>
                            <TooltipContent className="w-[300px]">
                              This connection does not have its own unique
                              wallet pubkey. Re-connect for additional privacy.
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}
                    </TableCell>
                  </TableRow>
                  {app.isolated && (
                    <TableRow>
                      <TableCell className="font-medium">Balance</TableCell>
                      <TableCell className="text-muted-foreground break-all">
                        {new Intl.NumberFormat().format(
                          Math.floor(app.balance / 1000)
                        )}{" "}
                        sats{" "}
                        <IsolatedAppTopupDialog appPubkey={app.appPubkey}>
                          <Button
                            size="sm"
                            variant="secondary"
                            className="ml-4"
                          >
                            Top Up
                          </Button>
                        </IsolatedAppTopupDialog>
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
                  <TableRow>
                    <TableCell className="font-medium">Created At</TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(app.createdAt).toString()}
                    </TableCell>
                  </TableRow>
                  {app.metadata && (
                    <TableRow>
                      <TableCell className="font-medium">Metadata</TableCell>
                      <TableCell className="text-muted-foreground whitespace-pre-wrap">
                        {JSON.stringify(app.metadata, null, 4)}
                      </TableCell>
                    </TableRow>
                  )}
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
                    {isEditingPermissions && (
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

                        {(app.isolated && !permissions.isolated) ||
                        (!app.scopes.includes("pay_invoice") &&
                          permissions.scopes.includes("pay_invoice")) ? (
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button type="button">Save</Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogTitle>
                                Confirm Update App
                              </AlertDialogTitle>
                              <AlertDialogDescription>
                                <div className="space-y-2">
                                  {app.isolated && !permissions.isolated ? (
                                    <p>
                                      Are you sure you wish to remove the
                                      <span className="font-bold">
                                        isolated
                                      </span>{" "}
                                      status from this connection?
                                    </p>
                                  ) : (
                                    <p>
                                      Are you sure you wish to give this
                                      connection{" "}
                                      <span className="font-bold">
                                        pay permissions
                                      </span>
                                      ?
                                    </p>
                                  )}
                                  <p className="text-amber-600 dark:text-amber-400 font-medium">
                                    ⚠️ Warning: This applies to all apps that
                                    have this connection secret. Only change
                                    this if you know it is safe to do so,
                                    otherwise you could potentially lose all
                                    funds
                                    {!!permissions.maxAmount &&
                                      " up to the specified budget"}
                                    {permissions.isolated &&
                                      " that are deposited into this isolated app"}
                                    .
                                  </p>
                                </div>
                              </AlertDialogDescription>
                              <AlertDialogFooter className="mt-5">
                                <AlertDialogCancel
                                  onClick={() => {
                                    window.location.reload();
                                  }}
                                >
                                  Cancel
                                </AlertDialogCancel>
                                <Button onClick={handleSave}>Save</Button>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        ) : (
                          <Button onClick={handleSave}>Save</Button>
                        )}
                      </div>
                    )}

                    {!isEditingPermissions && (
                      <>
                        <Button
                          variant="outline"
                          onClick={() =>
                            setIsEditingPermissions(!isEditingPermissions)
                          }
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
                readOnly={!isEditingPermissions}
                isNewConnection={false}
                budgetUsage={app.budgetUsage}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>
                <div className="flex flex-row justify-between items-center">
                  Transactions
                </div>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <TransactionsList appId={app.id} showReceiveButton={false} />
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}

export default ShowApp;
