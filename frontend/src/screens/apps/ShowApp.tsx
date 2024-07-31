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

import { DotsVerticalIcon } from "@radix-ui/react-icons";
import { Activity, ArrowDownCircle, ArrowUpCircle, Trash } from "lucide-react";
import { Bar, BarChart } from "recharts";
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
  CardTitle
} from "src/components/ui/card";
import { ChartConfig, ChartContainer } from "src/components/ui/chart";
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuTrigger } from "src/components/ui/dropdown-menu";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "src/components/ui/select";
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

  const chartData = [
    { month: "January", desktop: 186, mobile: 80 },
    { month: "February", desktop: 305, mobile: 200 },
    { month: "March", desktop: 237, mobile: 120 },
    { month: "April", desktop: 73, mobile: 190 },
    { month: "May", desktop: 209, mobile: 130 },
    { month: "June", desktop: 214, mobile: 140 },
  ]

  const chartConfig = {
    desktop: {
      label: "Desktop",
      color: "#000",
    },
    mobile: {
      label: "Mobile",
      color: "#000",
    },
  } satisfies ChartConfig

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
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <DotsVerticalIcon className="w-4 h-4" />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent className="w-56" align="end">
                    <DropdownMenuGroup>
                      <AlertDialogTrigger asChild>
                        <DropdownMenuItem>
                          <Trash className="mr-2 h-4 w-4" />
                          <span>Delete app</span>
                        </DropdownMenuItem>
                      </AlertDialogTrigger>
                    </DropdownMenuGroup>
                  </DropdownMenuContent>
                </DropdownMenu>
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

          <div className="flex justify-end items-center">
            <Select>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Last 30 days" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="light">This week</SelectItem>
                <SelectItem value="system">This month</SelectItem>
                <SelectItem value="dark">This year</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-3 gap-3">
            <Card className="flex flex-col">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  <div className="flex flex-row gap-2 items-center">
                    <ArrowDownCircle className="w-4 h-4" />
                    Incoming
                  </div>
                </CardTitle>
                <div className="text-2xl font-bold sensitive ph-no-capture">
                  123.456 sats
                </div>
              </CardHeader>
              <CardContent className="flex-grow">
                <ChartContainer config={chartConfig} className="h-40 w-full">
                  <BarChart data={chartData}>
                    <Bar dataKey="desktop" fill="var(--color-desktop)" radius={4} />
                  </BarChart>
                </ChartContainer>
              </CardContent>
            </Card>
            <Card className="flex flex-col">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  <div className="flex flex-row gap-2 items-center">
                    <ArrowUpCircle className="w-4 h-4" />
                    Outgoing
                  </div>
                </CardTitle>
                <div className="text-2xl font-bold sensitive ph-no-capture">
                  123.456 sats
                </div>
              </CardHeader>
              <CardContent className="flex-grow">
                <ChartContainer config={chartConfig} className="h-40 w-full">
                  <BarChart data={chartData}>
                    <Bar dataKey="mobile" fill="var(--color-mobile)" radius={4} />
                  </BarChart>
                </ChartContainer>

              </CardContent>
            </Card>
            <Card className="flex flex-col">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  <div className="flex flex-row gap-2 items-center">
                    <Activity className="w-4 h-4" />
                    Activity
                  </div>
                </CardTitle>
                <div className="text-2xl font-bold sensitive ph-no-capture">
                  123 usages
                </div>
              </CardHeader>
              <CardContent className="flex-grow">
                <ChartContainer config={chartConfig} className="h-40 w-full">
                  <BarChart data={chartData}>
                    <Bar dataKey="mobile" fill="var(--color-mobile)" radius={4} />
                  </BarChart>
                </ChartContainer>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Recent activity</CardTitle>
            </CardHeader>
            <CardContent>
              {/* <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>test</TableCell>
                    <TableCell>test</TableCell>
                    <TableCell>test</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  <TableRow>
                    <TableCell>test</TableCell>
                    <TableCell>test</TableCell>
                    <TableCell>test</TableCell>
                  </TableRow>
                </TableBody>
              </Table> */}
            </CardContent>
          </Card>

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
      </div >
    </>
  );
}

export default ShowApp;
