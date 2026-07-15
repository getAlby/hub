import { ArrowDownUpIcon } from "lucide-react";
import { useState } from "react";
import TransactionsList from "src/components/TransactionsList";
import { TransactionsListMenu } from "src/components/TransactionsListMenu";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  defaultTransactionFilters,
  type TransactionFilters,
} from "src/hooks/useTransactions";

export function AppTransactionList({ appId }: { appId: number }) {
  const [transactionFilters, setTransactionFilters] =
    useState<TransactionFilters>(defaultTransactionFilters);

  return (
    <Card>
      <CardHeader className="flex items-center justify-between">
        <CardTitle>Transactions</CardTitle>
        <TransactionsListMenu
          appId={appId}
          filters={transactionFilters}
          onFiltersChange={setTransactionFilters}
        />
      </CardHeader>
      <CardContent>
        <TransactionsList
          appId={appId}
          filters={transactionFilters}
          emptyIcon={ArrowDownUpIcon}
          emptyTitle="No transactions yet"
          emptyDescription="Payments made through this app will appear here."
          emptyVariant="none"
        />
      </CardContent>
    </Card>
  );
}
