import { ArrowDownUpIcon } from "lucide-react";
import TransactionsList from "src/components/TransactionsList";
import { TransactionsListMenu } from "src/components/TransactionsListMenu";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function AppTransactionList({ appId }: { appId: number }) {
  return (
    <Card>
      <CardHeader className="flex items-center justify-between">
        <CardTitle>Transactions</CardTitle>
        <TransactionsListMenu appId={appId} />
      </CardHeader>
      <CardContent>
        <TransactionsList
          appId={appId}
          emptyIcon={ArrowDownUpIcon}
          emptyTitle="No transactions yet"
          emptyDescription="Payments made through this app will appear here."
          emptyVariant="none"
        />
      </CardContent>
    </Card>
  );
}
