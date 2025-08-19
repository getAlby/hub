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
        <CardTitle>
          <div className="flex flex-row justify-between items-center">
            Transactions
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <TransactionsList appId={appId} showReceiveButton={false} />
      </CardContent>
    </Card>
  );
}
