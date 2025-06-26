import { AlertTriangleIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";

export function SpendingAlert({ className }: { className?: string }) {
  return (
    <Alert className={className}>
      <AlertTriangleIcon className="h-4 w-4" />
      <AlertTitle>Low spending capacity</AlertTitle>
      <AlertDescription>
        You won't be able to send payments until you{" "}
        <Link className="underline" to="/channels/outgoing">
          increase your spending capacity.
        </Link>
      </AlertDescription>
    </Alert>
  );
}
