import { AlertTriangleIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";

export function SwapUnavailableAlert() {
  return (
    <Alert variant="warning">
      <AlertTriangleIcon />
      <AlertTitle>Swaps are unavailable</AlertTitle>
      <AlertDescription>Swaps are temporarily disabled.</AlertDescription>
    </Alert>
  );
}
