import { CircleMinusIcon, CirclePlusIcon } from "lucide-react";
import React from "react";
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
import { cn, formatAmount, getBudgetRenewalLabel } from "src/lib/utils";
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
      setAllTransactions((current) => [
        ...current,
        ...transactionsResponse.transactions,
      ]);
      setPage((current) => current + 1);
    }
  }, [transactionsResponse?.transactions]);

  const totalSpent = allTransactions
    .filter((tx) => tx.type === "outgoing" && tx.state === "settled")
    .map((tx) => Math.floor(tx.amount / 1000))
    .reduce((a, b) => a + b, 0);

  const totalReceived = allTransactions
    .filter((tx) => tx.type === "incoming")
    .map((tx) => Math.floor(tx.amount / 1000))
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

  return (
    <>
      {app.isolated && (
        <div
          className={cn(
            "grid grid-cols-1 gap-2",
            app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID &&
              "lg:grid-cols-2"
          )}
        >
          <Card className="justify-between">
            <CardHeader>
              <CardTitle>Isolated Balance</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex justify-between items-end">
                <div>
                  <p className="font-medium text-2xl">
                    {new Intl.NumberFormat().format(
                      Math.floor(app.balance / 1000)
                    )}{" "}
                    sats
                  </p>
                  <FormattedFiatAmount amount={totalSpent} />
                </div>
                <div className="flex gap-2 items-center">
                  {app.balance > 0 && (
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
            </CardContent>
          </Card>
          {app.metadata?.app_store_app_id === SUBWALLET_APPSTORE_APP_ID && (
            <Card className="justify-between">
              <CardHeader>
                <CardTitle>Lightning Address</CardTitle>

                {!app.metadata.lud16 && (
                  <CardDescription>
                    <p className="text-muted-foreground">
                      Create a lightning address for this particular app
                      connection
                    </p>
                  </CardDescription>
                )}
              </CardHeader>
              <CardContent>
                {!app.metadata.lud16 && (
                  <div className="flex items-center gap-2">
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
                {app.metadata.lud16 && (
                  <div className="flex items-center justify-between">
                    <p className="font-semibold">{app.metadata.lud16}</p>
                    <LoadingButton
                      size="sm"
                      variant="outline"
                      loading={deletingLightningAddress}
                      onClick={deleteSubwalletLightningAddress}
                    >
                      Remove
                    </LoadingButton>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-2">
        <Card>
          <CardHeader>
            <CardTitle>Total Spent</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-medium text-2xl">
              {new Intl.NumberFormat().format(totalSpent)} sats
            </p>
            <FormattedFiatAmount amount={totalSpent} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Total Received</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-medium text-2xl">
              {new Intl.NumberFormat().format(totalReceived)} sats
            </p>
            <FormattedFiatAmount amount={totalReceived} />
          </CardContent>
        </Card>
      </div>

      {app.maxAmount > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Budget</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-row justify-between mb-2">
              <div>
                <p className="text-xs text-secondary-foreground font-medium">
                  Left in budget
                </p>
                <p className="text-xl font-medium">
                  {new Intl.NumberFormat().format(
                    app.maxAmount - app.budgetUsage
                  )}{" "}
                  sats
                </p>
                <FormattedFiatAmount amount={app.maxAmount - app.budgetUsage} />
              </div>
              <div>
                <p className="text-xs text-secondary-foreground font-medium">
                  Budget renewal
                </p>
                <p className="text-xl font-medium">
                  {formatAmount(app.maxAmount * 1000)} sats
                  {app.budgetRenewal !== "never" && (
                    <> / {getBudgetRenewalLabel(app.budgetRenewal)}</>
                  )}
                </p>
                <FormattedFiatAmount amount={app.maxAmount} />
              </div>
            </div>
            <Progress
              className="h-4"
              value={100 - (app.budgetUsage * 100) / app.maxAmount}
            />
          </CardContent>
        </Card>
      )}
    </>
  );
}
