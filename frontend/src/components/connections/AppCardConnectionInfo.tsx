import dayjs from "dayjs";
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
      {connection.maxAmount > 0 && (
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
              {connection.maxAmount && (
                <>{formatAmount(connection.maxAmount * 1000)} sats</>
              )}
            </div>
            <div>
              {connection.maxAmount > 0 &&
                connection.budgetRenewal !== "never" && (
                  <>Renews {connection.budgetRenewal}</>
                )}
            </div>
          </div>
          {connection.lastEventAt && (
            <div className="flex flex-row justify-between text-xs items-center text-muted-foreground">
              <div className="flex flex-row justify-between">
                <div>Last used:&nbsp;</div>
                <div>{dayjs(connection.lastEventAt).fromNow()}</div>
              </div>
            </div>
          )}
        </>
      )}
    </>
  );
}
