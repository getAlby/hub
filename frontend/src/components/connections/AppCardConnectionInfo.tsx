import dayjs from "dayjs";
import { CircleCheck, PlusCircle } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";
import { Progress } from "src/components/ui/progress";
import { formatAmount } from "src/lib/utils";
import { App } from "src/types";

type AppCardConnectionInfoProps = {
  connection: App;
};

export function AppCardConnectionInfo({
  connection,
}: AppCardConnectionInfoProps) {
  return (
    <>
      {connection.maxAmount > 0 ? (
        <>
          <div className="flex flex-row justify-between">
            <div className="mb-2">
              <p className="text-xs text-secondary-foreground font-medium">
                Left in budget
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
              {connection.lastEventAt && (
                <>Last used: {dayjs(connection.lastEventAt).fromNow()}</>
              )}
            </div>
            <div>
              {connection.maxAmount && (
                <>
                  {formatAmount(connection.maxAmount * 1000)} sats
                  {connection.budgetRenewal !== "never" && (
                    <> / {connection.budgetRenewal.slice(0, -2)}</>
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
          <div className="flex flex-row justify-end items-center">
            <Link to={`/apps/${connection.nostrPubkey}?edit=true`}>
              <Button variant="outline">
                <PlusCircle className="w-4 h-4 mr-2" />
                Set Budget
              </Button>
            </Link>
          </div>
        </>
      ) : (
        <>
          <div className="text-sm text-secondary-foreground font-medium w-full h-full flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
              <CircleCheck className="w-4 h-4" />
              Share wallet information
            </div>
            {connection.scopes.indexOf("make_invoice") > -1 && (
              <div className="flex flex-row items-center gap-2">
                <CircleCheck className="w-4 h-4" />
                Create Invoices
              </div>
            )}
          </div>
          <div className="flex flex-row justify-between items-center">
            <div className="flex flex-row justify-between text-xs items-center text-muted-foreground">
              {connection.lastEventAt && (
                <>Last used: {dayjs(connection.lastEventAt).fromNow()}</>
              )}
            </div>
            <Link to={`/apps/${connection.nostrPubkey}?edit=true`}>
              <Button variant="outline">
                <PlusCircle className="w-4 h-4 mr-2" />
                Enable Payments
              </Button>
            </Link>
          </div>
        </>
      )}
    </>
  );
}
