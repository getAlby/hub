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
  actions?: React.ReactNode;
  readonly?: boolean;
};

export default function AppCard({ app, actions, readonly = false }: Props) {
  const navigate = useNavigate();

  return (
    <Card
      className="flex flex-col group cursor-pointer"
      onClick={() => navigate(`/apps/${app.id}`)}
    >
      <CardHeader>
        <CardTitle className="relative min-w-0">
          <div className="flex flex-row items-center">
            {!actions && <AppCardNotice app={app} />}
            <AppAvatar className="w-10 h-10" app={app} />
            <div className="flex-1 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
              {app.name}
            </div>
            <div
              className="flex items-center gap-2"
              onClick={
                (e) =>
                  e.stopPropagation() /* stop the above navigation click handler */
              }
            >
              {actions}
            </div>
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col slashed-zero">
        <AppCardConnectionInfo connection={app} readonly={readonly} />
      </CardContent>
    </Card>
  );
}
