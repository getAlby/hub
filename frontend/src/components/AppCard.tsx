import gradientAvatar from "gradient-avatar";
import { Link } from "react-router-dom";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress";
import { App, NIP_47_PAY_INVOICE_METHOD } from "src/types";

type Props = {
  app: App;
  csrf?: string;
  onDelete: (nostrPubkey: string) => void;
};

export default function AppCard({ app, onDelete }: Props) {
  return (
    <>
      <Link to={`/connections/${app.nostrPubkey}`}>
        <Card>
          <CardHeader>
            <CardTitle>
              <div className="flex flex-row items-center">
                <div className="relative w-10 h-10 rounded-lg border">
                  <img
                    src={`data:image/svg+xml;base64,${btoa(
                      gradientAvatar(app.name)
                    )}`}
                    alt={app.name}
                    className="block w-full h-full rounded-lg p-1"
                  />
                  <span className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 text-white text-2xl font-medium capitalize">
                    {app.name.charAt(0)}
                  </span>
                </div>
                <h2 className="flex-1 font-semibold whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                  {app.name}
                </h2>
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {app.requestMethods?.includes(NIP_47_PAY_INVOICE_METHOD) ? (
              app.maxAmount > 0 ? (
                <>
                  <div className="mb-2">
                    <p className="text-sm">You've spent:</p>
                    <p className="text-xl font-medium">
                      {new Intl.NumberFormat().format(app.budgetUsage)} sats
                    </p>
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
          </CardContent>
        </Card>
      </Link>
    </>
  );
}
