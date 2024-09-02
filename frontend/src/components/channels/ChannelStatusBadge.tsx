import { InfoIcon } from "lucide-react";
import { Badge } from "src/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { Channel } from "src/types";

export function ChannelStatusBadge({ status }: { status: Channel["status"] }) {
  return status == "online" ? (
    <Badge variant="positive">Online</Badge>
  ) : status == "opening" ? (
    <Badge variant="outline">Opening</Badge>
  ) : status == "closing" ? (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <Badge variant="outline" className="text-muted-foreground">
            Closing&nbsp;
            <InfoIcon className="h-4 w-4 shrink-0" />
          </Badge>
        </TooltipTrigger>
        <TooltipContent className="w-[400px]">
          Any funds on your side of the channel at the time of closing will be
          returned to your savings account once the closing channel transaction
          has been broadcast. In case of force closures, this may take up to 2
          weeks.
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  ) : (
    <Badge variant="warning">Offline</Badge>
  );
}
