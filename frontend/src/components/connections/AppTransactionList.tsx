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
        <TransactionsList appId={appId} showReceiveButton={false} />
      </CardContent>
    </Card>
  );
}
