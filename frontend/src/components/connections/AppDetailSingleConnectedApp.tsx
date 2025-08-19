import { AppTransactionList } from "src/components/connections/AppTransactionList";
import { AppUsage } from "src/components/connections/AppUsage";
import Permissions from "src/components/Permissions";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useCapabilities } from "src/hooks/useCapabilities";
import { App } from "src/types";

export function AppDetailSingleConnectedApp({ app }: { app: App }) {
  const { data: capabilities } = useCapabilities();
  if (!capabilities) {
    return null;
  }
  return (
    <>
      <AppUsage app={app} />
      <Card>
        <CardHeader>
          <CardTitle>Permissions</CardTitle>
        </CardHeader>
        <CardContent>
          <Permissions
            capabilities={capabilities}
            permissions={{
              scopes: app.scopes,
              maxAmount: app.maxAmount,
              budgetRenewal: app.budgetRenewal,
              expiresAt: app.expiresAt ? new Date(app.expiresAt) : undefined,
              isolated: app.isolated,
            }}
            readOnly
            isNewConnection={false}
            budgetUsage={app.budgetUsage}
          />
        </CardContent>
      </Card>
      <AppTransactionList appId={app.id} />
    </>
  );
}
