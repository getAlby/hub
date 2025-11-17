import dayjs from "dayjs";
import { BrickWallIcon, CircleCheckIcon, PlusCircleIcon } from "lucide-react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { LinkButton } from "src/components/ui/custom/link-button";
import { Progress } from "src/components/ui/progress";
import { SUBWALLET_APPSTORE_APP_ID } from "src/constants";
import { getBudgetRenewalLabel } from "src/lib/utils";
import { App } from "src/types";

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
              {connection.lastUsedAt
                ? dayjs(connection.lastUsedAt).fromNow()
                : "Never"}
            </div>
            <div className="flex flex-col items-end justify-end">
              <p>Balance</p>
              <p className="text-xl font-medium sensitive">
                <FormattedBitcoinAmount amount={connection.balance} />
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
              <p className="text-xl font-medium sensitive">
                <FormattedBitcoinAmount
                  amount={
                    (connection.maxAmount - connection.budgetUsage) * 1000
                  }
                />
              </p>
            </div>
          </div>
          <Progress
            className="h-4"
            value={100 - (connection.budgetUsage * 100) / connection.maxAmount}
          />
          <div className="flex flex-row justify-between text-xs items-center text-muted-foreground mt-2">
            <div>
              Last used:{" "}
              {connection.lastUsedAt
                ? dayjs(connection.lastUsedAt).fromNow()
                : "Never"}
            </div>
            <div>
              {connection.maxAmount && (
                <>
                  <FormattedBitcoinAmount
                    amount={connection.maxAmount * 1000}
                  />
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
              <p className="text-xl font-medium sensitive">
                <FormattedBitcoinAmount
                  amount={connection.budgetUsage * 1000}
                />
              </p>
            </div>
          </div>
          <div className="flex flex-row justify-between items-center">
            <div className="text-muted-foreground text-xs">
              Last used:{" "}
              {connection.lastUsedAt
                ? dayjs(connection.lastUsedAt).fromNow()
                : "Never"}
            </div>
            {!readonly && (
              <LinkButton
                to={`/apps/${connection.id}?edit=true`}
                variant="outline"
              >
                <PlusCircleIcon />
                Set Budget
              </LinkButton>
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
              {connection.lastUsedAt
                ? dayjs(connection.lastUsedAt).fromNow()
                : "Never"}
            </div>
            {!readonly && (
              <LinkButton
                to={`/apps/${connection.id}?edit=true`}
                variant="outline"
              >
                <PlusCircleIcon />
                Enable Payments
              </LinkButton>
            )}
          </div>
        </>
      )}
    </>
  );
}
