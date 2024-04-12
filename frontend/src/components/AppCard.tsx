import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { Link } from "react-router-dom";
import AppAvatar from "src/components/AppAvatar";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress";
import { App, NIP_47_PAY_INVOICE_METHOD } from "src/types";

dayjs.extend(relativeTime);

type Props = {
  app: App;
  csrf?: string;
};

export default function AppCard({ app }: Props) {
  return (
    <>
      <Link to={`/apps/${app.nostrPubkey}`}>
        <Card>
          <CardHeader>
            <CardTitle>
              <div className="flex flex-row items-center">
                <AppAvatar className="w-10 h-10" appName={app.name} />
                <h2 className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                  {app.name}
                </h2>
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {app.requestMethods?.includes(NIP_47_PAY_INVOICE_METHOD) ? (
              app.maxAmount > 0 ? (
                <>
                  <div className="flex flex-row justify-between">
                    <div className="mb-2">
                      <p className="text-xs text-secondary-foreground font-medium">
                        You've spent
                      </p>
                      <p className="text-xl font-medium">
                        {new Intl.NumberFormat().format(app.budgetUsage)} sats
                      </p>
                    </div>
                    <div className="text-right">
                      {" "}
                      <p className="text-xs text-secondary-foreground font-medium">
                        Left in budget
                      </p>
                      <p className="text-xl font-medium text-muted-foreground">
                        {new Intl.NumberFormat().format(
                          app.maxAmount - app.budgetUsage
                        )}{" "}
                        sats
                      </p>
                    </div>
                  </div>
                  <Progress
                    className="h-4"
                    value={(app.budgetUsage * 100) / app.maxAmount}
                  />
                </>
              ) : (
                "No limits!"
              )
            ) : (
              "Payments disabled."
            )}
            <div className="grid gap-2 mt-5 text-muted-foreground text-sm">
              <div className="flex flex-row justify-between">
                <div>Budget</div>
                <div>
                  {app.maxAmount > 0 ? (
                    <>
                      {new Intl.NumberFormat().format(app.maxAmount)} sats /{" "}
                      {app.budgetRenewal}
                    </>
                  ) : (
                    "Not set"
                  )}
                </div>
              </div>
              <div className="flex flex-row justify-between">
                <div>Expires on</div>
                <div>
                  {app.expiresAt ? dayjs(app.expiresAt).fromNow() : "Never"}
                </div>
              </div>
              <div className="flex flex-row justify-between">
                <div>Last used</div>
                <div>
                  {app.lastEventAt ? dayjs(app.lastEventAt).fromNow() : "Never"}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </Link>
    </>
  );
}
