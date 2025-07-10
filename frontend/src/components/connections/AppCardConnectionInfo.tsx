import dayjs from "dayjs";
import { BrickWallIcon, CircleCheckIcon, PlusCircleIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";
import { Progress } from "src/components/ui/progress";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { formatAmount } from "src/lib/utils";
import { App, BudgetRenewalType } from "src/types";

type AppCardConnectionInfoProps = {
  connection: App;
  budgetRemainingText?: string | React.ReactNode;
  readonly?: boolean;
};

export function AppCardConnectionInfo({
  connection,
  budgetRemainingText = "Left in budget",
  readonly = false,
}: AppCardConnectionInfoProps) {
  function getBudgetRenewalLabel(renewalType: BudgetRenewalType): string {
    switch (renewalType) {
      case "daily":
        return "day";
      case "weekly":
        return "week";
      case "monthly":
        return "month";
      case "yearly":
        return "year";
      case "never":
        return "never";
      case "":
        return "";
    }
  }

  return (
    <>
      {connection.isolated ? (
        <>
          <div className="text-sm text-secondary-foreground font-medium w-full h-full flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
              <BrickWallIcon className="size-4" />

              {connection.metadata?.app_store_app_id ===
              SUBWALLET_APPSTORE_APP_ID
                ? "Sub-wallet"
                : "Isolated App"}
            </div>
          </div>
          <div className="flex flex-row justify-between text-xs items-end mt-2">
            <div className="text-muted-foreground">
              Last used:{" "}
              {connection.lastEventAt
                ? dayjs(connection.lastEventAt).fromNow()
                : "Never"}
            </div>
            <div className="flex flex-col items-end justify-end">
              <p>Balance</p>
              <p className="text-xl font-medium">
                {new Intl.NumberFormat().format(
                  Math.floor(connection.balance / 1000)
                )}{" "}
                sats
              </p>
            </div>
          </div>
        </>
      ) : connection.maxAmount > 0 ? (
        <>
          <div className="flex flex-row justify-between">
            <div className="mb-2">
              <p className="text-xs text-secondary-foreground font-medium">
                {budgetRemainingText}
              </p>
              <p className="text-xl font-medium">
                {new Intl.NumberFormat().format(
                  connection.maxAmount - connection.budgetUsage
                )}{" "}
                sats
              </p>
            </div>
          </div>
          <Progress
            className="h-4"
            value={(connection.budgetUsage * 100) / connection.maxAmount}
          />
          <div className="flex flex-row justify-between text-xs items-center text-muted-foreground mt-2">
            <div>
              Last used:{" "}
              {connection.lastEventAt
                ? dayjs(connection.lastEventAt).fromNow()
                : "Never"}
            </div>
            <div>
              {connection.maxAmount && (
                <>
                  {formatAmount(connection.maxAmount * 1000)} sats
                  {connection.budgetRenewal !== "never" && (
                    <> / {getBudgetRenewalLabel(connection.budgetRenewal)}</>
                  )}
                </>
              )}
            </div>
          </div>
        </>
      ) : connection.scopes.indexOf("pay_invoice") > -1 ? (
        <>
          <div className="flex flex-row justify-between">
            <div className="mb-2">
              <p className="text-xs text-secondary-foreground font-medium">
                You've spent
              </p>
              <p className="text-xl font-medium">
                {new Intl.NumberFormat().format(connection.budgetUsage)} sats
              </p>
            </div>
          </div>
          <div className="flex flex-row justify-between items-center">
            <div className="text-muted-foreground text-xs">
              Last used:{" "}
              {connection.lastEventAt
                ? dayjs(connection.lastEventAt).fromNow()
                : "Never"}
            </div>
            {!readonly && (
              <Link to={`/apps/${connection.appPubkey}?edit=true`}>
                <Button variant="outline">
                  <PlusCircleIcon className="size-4 mr-2" />
                  Set Budget
                </Button>
              </Link>
            )}
          </div>
        </>
      ) : (
        <>
          <div className="text-sm text-secondary-foreground font-medium w-full h-full flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
              <CircleCheckIcon className="size-4" />
              Share wallet information
            </div>
            {connection.scopes.indexOf("make_invoice") > -1 && (
              <div className="flex flex-row items-center gap-2">
                <CircleCheckIcon className="size-4" />
                Receive payments
              </div>
            )}
            {connection.scopes.indexOf("list_transactions") > -1 && (
              <div className="flex flex-row items-center gap-2">
                <CircleCheckIcon className="size-4" />
                Read transaction history
              </div>
            )}
          </div>
          <div className="flex flex-row justify-between items-center">
            <div className="flex flex-row justify-between text-xs items-center text-muted-foreground">
              Last used:{" "}
              {connection.lastEventAt
                ? dayjs(connection.lastEventAt).fromNow()
                : "Never"}
            </div>
            {!readonly && (
              <Link
                to={`/apps/${connection.appPubkey}?edit=true`}
                onClick={(e) => e.stopPropagation()}
              >
                <Button variant="outline">
                  <PlusCircleIcon className="size-4 mr-2" />
                  Enable Payments
                </Button>
              </Link>
            )}
          </div>
        </>
      )}
    </>
  );
}
