import { AlertTriangleIcon } from "lucide-react";
import { Link } from "react-router";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";

export default function LowReceivingCapacityAlert() {
  return (
    <Alert variant="warning">
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>You need more receiving capacity</AlertTitle>
      <AlertDescription className="inline">
        This wallet cannot receive larger payments right now. Add receiving
        capacity,{" "}
        <Link className="underline" to="/wallet/send">
          spend
        </Link>
        ,{" "}
        <Link className="underline" to="/wallet/swap?type=out">
          swap out funds
        </Link>
        , or{" "}
        <Link className="underline" to="/channels/incoming">
          open an incoming channel.
        </Link>
      </AlertDescription>
    </Alert>
  );
}
