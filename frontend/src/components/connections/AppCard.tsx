import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { useNavigate } from "react-router-dom";
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
};

export default function AppCard({ app }: Props) {
  const navigate = useNavigate();

  return (
    <Card
      className="flex flex-col group cursor-pointer"
      onClick={() => navigate(`/apps/${app.nostrPubkey}`)}
    >
      <CardHeader>
        <CardTitle className="relative">
          <AppCardNotice app={app} />
          <div className="flex flex-row items-center">
            <AppAvatar className="w-10 h-10" app={app} />
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
  );
}
