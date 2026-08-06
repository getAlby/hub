import { TriangleAlertIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useChannels } from "src/hooks/useChannels";
import { sendEvent } from "src/utils/sendEvent";

export function PaymentFailedAlert({
  invoice,
  errorMessage,
}: {
  invoice: string;
  errorMessage: string;
}) {
  const [sendingDetailsToAlby, setSendingDetailsToAlby] = React.useState(false);
  const { data: channels } = useChannels();

  async function sendDetailsToAlby() {
    setSendingDetailsToAlby(true);
    await sendEvent("payment_failed_details", {
      invoice,
      errorMessage,
      channels,
    });
    toast("Thanks for improving Alby Hub.");
    setSendingDetailsToAlby(false);
  }

  return (
    <Alert>
      <TriangleAlertIcon className="h-4 w-4" />
      <AlertTitle>Payment was not sent</AlertTitle>
      <AlertDescription>
        <p>
          Alby Hub could not complete this payment. No sats were sent unless you
          see it in your transactions.
        </p>
        <div className="flex flex-wrap gap-2 mt-2">
          <ExternalLinkButton
            to="https://guides.getalby.com/user-guide/alby-hub/faq/what-to-do-if-i-cannot-send-a-payment"
            size={"sm"}
          >
            View Payments Guide
          </ExternalLinkButton>
          <LoadingButton
            size={"sm"}
            variant="secondary"
            loading={sendingDetailsToAlby}
            onClick={sendDetailsToAlby}
          >
            Send Details to Alby
          </LoadingButton>
        </div>
      </AlertDescription>
    </Alert>
  );
}
