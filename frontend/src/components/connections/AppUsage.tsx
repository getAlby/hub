import {
  CircleMinusIcon,
  CirclePlusIcon,
  CopyIcon,
  Trash2Icon,
} from "lucide-react";
import React from "react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { IsolatedAppDrawDownDialog } from "src/components/IsolatedAppDrawDownDialog";
import { IsolatedAppTopupDialog } from "src/components/IsolatedAppTopupDialog";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { InputWithAdornment } from "src/components/ui/custom/input-with-adornment";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Progress } from "src/components/ui/progress";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCreateLightningAddress } from "src/hooks/useCreateLightningAddress";
import { useDeleteLightningAddress } from "src/hooks/useDeleteLightningAddress";
import { useTransactions } from "src/hooks/useTransactions";
import { copyToClipboard } from "src/lib/clipboard";
import { cn, getBudgetRenewalLabel } from "src/lib/utils";
import { App, Transaction } from "src/types";

export function AppUsage({ app }: { app: App }) {
  const [page, setPage] = React.useState(1);
  const { data: transactionsResponse } = useTransactions(
    app.id,
    false,
    100,
    page
  );
  const [allTransactions, setAllTransactions] = React.useState<Transaction[]>(
    []
  );
  React.useEffect(() => {
    if (transactionsResponse?.transactions.length) {
      setAllTransactions((current) =>
        [...current, ...transactionsResponse.transactions].filter(
          (v, i, a) => a.findIndex((t) => t.paymentHash === v.paymentHash) === i // remove duplicates
        )
      );
      setPage((current) => current + 1);
    }
  }, [transactionsResponse?.transactions]);

  const totalSpentSat = allTransactions
    .filter((tx) => tx.type === "outgoing" && tx.state === "settled")
    .map((tx) => Math.floor((tx.amountMsat + tx.feesPaidMsat) / 1000))
    .reduce((a, b) => a + b, 0);

  const totalReceivedSat = allTransactions
    .filter((tx) => tx.type === "incoming")
    .map((tx) => tx.amountSat)
    .reduce((a, b) => a + b, 0);

  const { data: albyMe } = useAlbyMe();
  const [intendedLightningAddress, setIntendedLightningAddress] =
    React.useState("");
  const { createLightningAddress, creatingLightningAddress } =
    useCreateLightningAddress(app.id);
  const {
    deleteLightningAddress: deleteSubwalletLightningAddress,
    deletingLightningAddress,
  } = useDeleteLightningAddress(app.id);

  const isSubwallet =
    app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID;

  return (
    <>
      <Card className="slashed-zero gap-4">
        <CardHeader>
          <CardTitle>Usage</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-5">
          <div
            className={cn(
              "grid grid-cols-1 gap-6 sm:gap-4",
              app.isolated ? "sm:grid-cols-3" : "sm:grid-cols-2"
            )}
          >
            {app.isolated && (
              <div className="flex flex-col">
                <p className="text-xs text-secondary-foreground font-medium">
                  Isolated Balance
                </p>
                <p className="text-xl font-medium">
                  <FormattedBitcoinAmount amountMsat={app.balanceMsat} />
                </p>
                <FormattedFiatAmount
                  amountSat={Math.floor(app.balanceMsat / 1000)}
                />
                <div className="flex gap-2 items-center mt-3">
                  {app.balanceMsat > 0 && (
                    <IsolatedAppDrawDownDialog appId={app.id}>
                      <Button size="sm" variant="outline">
                        <CircleMinusIcon />
                        Decrease
                      </Button>
                    </IsolatedAppDrawDownDialog>
                  )}
                  <IsolatedAppTopupDialog appId={app.id}>
                    <Button size="sm" variant="outline">
                      <CirclePlusIcon />
                      Increase
                    </Button>
                  </IsolatedAppTopupDialog>
                </div>
              </div>
            )}
            <div className="flex flex-col">
              <p className="text-xs text-secondary-foreground font-medium">
                Total Spent
              </p>
              <p className="text-xl font-medium">
                <FormattedBitcoinAmount amountMsat={totalSpentSat * 1000} />
              </p>
              <FormattedFiatAmount amountSat={totalSpentSat} />
            </div>
            <div className="flex flex-col">
              <p className="text-xs text-secondary-foreground font-medium">
                Total Received
              </p>
              <p className="text-xl font-medium">
                <FormattedBitcoinAmount amountMsat={totalReceivedSat * 1000} />
              </p>
              <FormattedFiatAmount amountSat={totalReceivedSat} />
            </div>
          </div>

          {app.maxAmountSat > 0 && (
            <div className="flex flex-col gap-2 border-t pt-5">
              <div className="flex flex-row justify-between">
                <div>
                  <p className="text-xs text-secondary-foreground font-medium">
                    Left in budget
                  </p>
                  <p className="text-xl font-medium">
                    <FormattedBitcoinAmount
                      amountMsat={app.maxAmountMsat - app.budgetUsageMsat}
                    />
                  </p>
                  <FormattedFiatAmount
                    amountSat={app.maxAmountSat - app.budgetUsageSat}
                  />
                </div>
                <div className="text-right">
                  <p className="text-xs text-secondary-foreground font-medium">
                    Budget renewal
                  </p>
                  <p className="text-xl font-medium">
                    <FormattedBitcoinAmount amountMsat={app.maxAmountMsat} />
                    {app.budgetRenewal !== "never" && (
                      <> / {getBudgetRenewalLabel(app.budgetRenewal)}</>
                    )}
                  </p>
                  <FormattedFiatAmount amountSat={app.maxAmountSat} />
                </div>
              </div>
              <Progress
                className="h-2 mt-1"
                value={100 - (app.budgetUsageSat * 100) / app.maxAmountSat}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {app.isolated && isSubwallet && (
        <Card>
          <CardHeader>
            <CardTitle>Lightning Address</CardTitle>

            {!app.metadata?.lud16 && (
              <CardDescription>
                Create a lightning address for this particular app connection
              </CardDescription>
            )}
          </CardHeader>
          <CardContent>
            {!app.metadata?.lud16 && (
              <div className="flex items-center gap-2">
                <InputWithAdornment
                  type="text"
                  value={intendedLightningAddress}
                  onChange={(e) => setIntendedLightningAddress(e.target.value)}
                  required
                  autoComplete="off"
                  endAdornment={
                    <span className="mr-1 text-muted-foreground text-xs">
                      @getalby.com
                    </span>
                  }
                  className="flex-1"
                />
                {!albyMe?.subscription.plan_code ? (
                  <UpgradeDialog>
                    <Button className="shrink-0" variant="secondary">
                      Create
                    </Button>
                  </UpgradeDialog>
                ) : (
                  <LoadingButton
                    className="shrink-0"
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
            {app.metadata?.lud16 && (
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                <p className="font-semibold break-all min-w-0 flex-1">
                  {app.metadata.lud16}
                </p>
                <div className="flex items-center gap-2 shrink-0">
                  <Button
                    size="sm"
                    onClick={() => copyToClipboard(app.metadata?.lud16 || "")}
                    variant="outline"
                  >
                    <CopyIcon />
                    Copy
                  </Button>
                  <LoadingButton
                    size="sm"
                    variant="outline"
                    loading={deletingLightningAddress}
                    onClick={deleteSubwalletLightningAddress}
                  >
                    <Trash2Icon />
                    Remove
                  </LoadingButton>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </>
  );
}
