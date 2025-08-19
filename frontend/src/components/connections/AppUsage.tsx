import React from "react";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress";
import { useTransactions } from "src/hooks/useTransactions";
import { formatAmount, getBudgetRenewalLabel } from "src/lib/utils";
import { App, Transaction } from "src/types";

export function AppUsage({ app }: { app: App }) {
  const [page, setPage] = React.useState(1);
  const { data: transactionsResponse } = useTransactions(
    app.id,
    false,
    1,
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

  return (
    <>
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
