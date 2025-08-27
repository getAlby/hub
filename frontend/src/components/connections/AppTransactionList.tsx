import TransactionsList from "src/components/TransactionsList";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function AppTransactionList({ appId }: { appId: number }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Transactions</CardTitle>
      </CardHeader>
      <CardContent>
        <TransactionsList appId={appId} showReceiveButton={false} />
      </CardContent>
    </Card>
  );
}
