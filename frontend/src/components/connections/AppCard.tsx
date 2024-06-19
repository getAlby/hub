import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { Link } from "react-router-dom";
import AppAvatar from "src/components/AppAvatar";
import { AppCardConnectionInfo } from "src/components/connections/AppCardConnectionInfo";
import { AppCardNotice } from "src/components/connections/AppCardNotice";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { App } from "src/types";

dayjs.extend(relativeTime);

type Props = {
  app: App;
  csrf?: string;
};

export default function AppCard({ app }: Props) {
  return (
    <>
      <Link className="h-full" to={`/apps/${app.nostrPubkey}`}>
        <Card className="h-full flex flex-col group">
          <CardHeader>
            <CardTitle className="relative">
              <AppCardNotice app={app} />
              <div className="flex flex-row items-center">
                <AppAvatar className="w-10 h-10" appName={app.name} />
                <div className="flex-1 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                  {app.name}
                </div>
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent className="flex-1 flex flex-col">
            <AppCardConnectionInfo connection={app} />
          </CardContent>
        </Card>
      </Link>
    </>
  );
}
