import { AlertTriangleIcon } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert.tsx";

export default function LowReceivingCapacityAlert() {
  return (
    <Alert variant="warning">
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>Low receiving capacity</AlertTitle>
      <AlertDescription className="inline">
        You likely won't be able to receive payments until you{" "}
        <Link className="underline" to="/wallet/send">
          spend
        </Link>
        ,{" "}
        <Link className="underline" to="/wallet/swap?type=out">
          swap out funds
        </Link>
        , or{" "}
        <Link className="underline" to="/channels/incoming">
          increase your receiving capacity.
        </Link>
      </AlertDescription>
    </Alert>
  );
}
