import { AlertTriangleIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";

export function OnchainTransactionCheckFailedAlert() {
  return (
    <Alert variant="warning">
      <AlertTriangleIcon />
      <AlertTitle>Could not check transaction</AlertTitle>
      <AlertDescription>
        Alby Hub could not find this transaction on mempool. View it on mempool
        to check its status before sending again.
      </AlertDescription>
    </Alert>
  );
}
