import { AlertCircleIcon, ExternalLinkIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import { getAppStoreApp } from "src/components/connections/SuggestedAppData";
import { IsolatedAppDrawDownDialog } from "src/components/IsolatedAppDrawDownDialog";
import { IsolatedAppTopupDialog } from "src/components/IsolatedAppTopupDialog";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCreateLightningAddress } from "src/hooks/useCreateLightningAddress";
import { useDeleteLightningAddress } from "src/hooks/useDeleteLightningAddress";
import { App } from "src/types";

export function ConnectionSummary({ app }: { app: App }) {
  const { data: albyMe } = useAlbyMe();
  const appStoreApp = getAppStoreApp(app);
  const [intendedLightningAddress, setIntendedLightningAddress] =
    React.useState("");
  const { createLightningAddress, creatingLightningAddress } =
    useCreateLightningAddress(app.id);
  const {
    deleteLightningAddress: deleteSubwalletLightningAddress,
    deletingLightningAddress,
  } = useDeleteLightningAddress(app.id);
  return (
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
              <TableCell className="font-medium">App Public Key</TableCell>
              <TableCell className="text-muted-foreground break-all">
                {app.appPubkey}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell className="font-medium">Wallet Public Key</TableCell>
              <TableCell className="text-muted-foreground flex items-center">
                <span className="break-all">{app.walletPubkey}</span>
                {!app.uniqueWalletPubkey && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <AlertCircleIcon className="w-3 h-3 ml-2 flex-shrink-0" />
                      </TooltipTrigger>
                      <TooltipContent className="w-[300px]">
                        This connection does not have its own unique wallet
                        pubkey. Re-connect for additional privacy.
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
                  <IsolatedAppTopupDialog appId={app.id}>
                    <Button size="sm" variant="secondary" className="ml-4">
                      Top Up
                    </Button>
                  </IsolatedAppTopupDialog>{" "}
                  {app.balance > 0 && (
                    <IsolatedAppDrawDownDialog appId={app.id}>
                      <Button size="sm" variant="secondary" className="ml-4">
                        Draw Down
                      </Button>
                    </IsolatedAppDrawDownDialog>
                  )}
                </TableCell>
              </TableRow>
            )}
            {app.isolated &&
              app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID && (
                <TableRow>
                  <TableCell className="font-medium">
                    Lightning Address
                  </TableCell>
                  <TableCell className="text-muted-foreground break-all">
                    {app.metadata.lud16}
                    {!app.metadata.lud16 && (
                      <div className="max-w-96 flex items-center gap-2">
                        <InputWithAdornment
                          type="text"
                          value={intendedLightningAddress}
                          onChange={(e) =>
                            setIntendedLightningAddress(e.target.value)
                          }
                          required
                          autoComplete="off"
                          endAdornment={
                            <span className="mr-1 text-muted-foreground text-xs">
                              @getalby.com
                            </span>
                          }
                        />
                        {!albyMe?.subscription.plan_code ? (
                          <UpgradeDialog>
                            <Button
                              className="shrink-0"
                              size="lg"
                              variant="secondary"
                            >
                              Create
                            </Button>
                          </UpgradeDialog>
                        ) : (
                          <LoadingButton
                            className="shrink-0"
                            size="lg"
                            variant="secondary"
                            loading={creatingLightningAddress}
                            onClick={() =>
                              createLightningAddress(intendedLightningAddress)
                            }
                          >
                            Create
                          </LoadingButton>
                        )}
                      </div>
                    )}
                    {app.metadata.lud16 && (
                      <LoadingButton
                        size="sm"
                        variant="destructive"
                        className="ml-4"
                        loading={deletingLightningAddress}
                        onClick={deleteSubwalletLightningAddress}
                      >
                        Remove
                      </LoadingButton>
                    )}
                  </TableCell>
                </TableRow>
              )}
            <TableRow>
              <TableCell className="font-medium">Last used</TableCell>
              <TableCell className="text-muted-foreground">
                {app.lastUsedAt ? new Date(app.lastUsedAt).toString() : "Never"}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell className="font-medium">Expires At</TableCell>
              <TableCell className="text-muted-foreground">
                {app.expiresAt ? new Date(app.expiresAt).toString() : "Never"}
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
            {appStoreApp && (
              <TableRow>
                <TableCell className="font-medium">App</TableCell>
                <TableCell className="text-muted-foreground">
                  <Link
                    to={
                      appStoreApp.internal
                        ? `/internal-apps/${appStoreApp.id}`
                        : `/appstore/${appStoreApp.id}`
                    }
                    className="flex items-center gap-2"
                  >
                    <ExternalLinkIcon className="size-4" /> {appStoreApp.title}
                  </Link>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
