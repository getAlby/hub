import dayjs from "dayjs";
import { ChevronRightIcon } from "lucide-react";
import { Link } from "react-router";
import AppAvatar from "src/components/AppAvatar";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useApps } from "src/hooks/useApps";
import { useTransactions } from "src/hooks/useTransactions";
import { cn, getAppDisplayName } from "src/lib/utils";
import { App } from "src/types";

export function LatestUsedAppsWidget() {
  const { data: appsData } = useApps(
    3,
    undefined,
    undefined,
    "last_settled_transaction"
  );
  const apps = appsData?.apps;
  const usedApps = apps?.filter((x) => x.lastSettledTransactionAt);

  if (!usedApps?.length) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div>Recently Used Apps</div>
          <LinkButton to="/apps?tab=connected-apps" variant="ghost" size="sm">
            See All
          </LinkButton>
        </CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-4">
        {usedApps
          .sort(
            (a, b) =>
              new Date(b.lastSettledTransactionAt ?? 0).getTime() -
              new Date(a.lastSettledTransactionAt ?? 0).getTime()
          )
          .map((app) => (
            <RecentlyUsedAppRow key={app.id} app={app} />
          ))}
      </CardContent>
    </Card>
  );
}

function RecentlyUsedAppRow({ app }: { app: App }) {
  const { data: transactionsData } = useTransactions(app.id, false, 1, 1);
  const latestTransaction = transactionsData?.transactions[0];

  return (
    <Link to={`/apps/${app.id}`} className="group">
      <div className="flex items-center w-full gap-4">
        <AppAvatar app={app} className="w-14 h-14 rounded-lg" />
        <p className="text-sm font-medium flex-1 truncate">
          {getAppDisplayName(app.name)}
        </p>
        <div className="flex flex-col items-end">
          {latestTransaction && (
            <span
              className={cn(
                "text-sm font-medium",
                latestTransaction.type === "incoming" &&
                  "text-green-600 dark:text-emerald-500"
              )}
            >
              {latestTransaction.type === "outgoing" ? "-" : "+"}
              <FormattedBitcoinAmount
                amountMsat={latestTransaction.amountMsat}
              />
            </span>
          )}
          <span className="text-xs text-muted-foreground">
            {app.lastSettledTransactionAt
              ? dayjs(app.lastSettledTransactionAt).fromNow()
              : "never"}
          </span>
        </div>
        <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
      </div>
    </Link>
  );
}
