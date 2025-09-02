import { TriangleAlertIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useChannels } from "src/hooks/useChannels";
import { request } from "src/utils/request";

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
    try {
      await request(`/api/event`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          event: "payment_failed_details",
          properties: {
            invoice,
            errorMessage,
            channels,
          },
        }),
      });
      toast("Thanks for improving Alby Hub.");
    } catch (error) {
      console.error(error);
    }
    setSendingDetailsToAlby(false);
  }

  return (
    <Alert>
      <TriangleAlertIcon className="h-4 w-4" />
      <AlertTitle>Payment Failed</AlertTitle>
      <AlertDescription>
        <p>
          Try the payment again, read our payments guide, and optionally send
          details about the failed payment to help improve Alby Hub.
        </p>
        <div className="flex gap-2 mt-2">
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
