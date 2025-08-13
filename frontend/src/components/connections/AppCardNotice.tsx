import dayjs from "dayjs";
import isBetween from "dayjs/plugin/isBetween"; // Add this line
import { CalendarClockIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { Badge } from "src/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { App } from "src/types";

dayjs.extend(isBetween); // Extend dayjs with the isBetween plugin

type AppCardNoticeProps = {
  app: App;
};

export function AppCardNotice({ app }: AppCardNoticeProps) {
  const now = dayjs();
  const expiresAt = dayjs(app.expiresAt);
  const isExpired = expiresAt.isBefore(now);
  const expiresSoon = expiresAt.isBetween(now, now.add(7, "days"));

  return (
    <div className="absolute top-0 right-0">
      {app.expiresAt ? (
        isExpired ? (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Link to={`/apps/${app.id}`}>
                  <Badge variant="destructive">
                    <CalendarClockIcon className="w-3 h-3 mr-2" />
                    Expired
                  </Badge>
                </Link>
              </TooltipTrigger>
              <TooltipContent>Expired {expiresAt.fromNow()}</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        ) : expiresSoon ? (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Link to={`/apps/${app.id}`}>
                  <Badge variant="outline">
                    <CalendarClockIcon className="w-3 h-3 mr-2" />
                    Expires Soon
                  </Badge>
                </Link>
              </TooltipTrigger>
              <TooltipContent>Expires {expiresAt.fromNow()}</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        ) : null
      ) : null}
    </div>
  );
}
