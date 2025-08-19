import React from "react";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useTransactions } from "src/hooks/useTransactions";
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
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
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
  );
}
