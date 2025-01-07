import React, { useEffect } from "react";
import AppHeader from "src/components/AppHeader";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { useToast } from "src/components/ui/use-toast";

export function Bitrefill() {
  const { toast } = useToast();
  const [open, setOpen] = React.useState(false);
  const [invoice, setInvoice] = React.useState("");

  useEffect(() => {
    const handleIframeMessage = (event: MessageEvent) => {
      if (event.origin === "https://embed.bitrefill.com") {
        console.log("💯", event.data);
        // Handle the message event data as needed

        const parsedData = JSON.parse(event.data);
        console.log(parsedData);

        if (parsedData.event === "payment_intent") {
          setOpen(true);
          setInvoice(parsedData.paymentAddress);
        }
      }
    };

    window.addEventListener("message", handleIframeMessage);

    return () => {
      window.removeEventListener("message", handleIframeMessage);
    };
  }, []);

  function confirmPayment() {
    setOpen(false);
    setInvoice("");
    // TODO: execute payment
    toast({ title: "Succesfully paid" });
  }

  // TODO:
  //  - ref param
  //  - Add email?
  //  - Refund address
  //  - parse invoice

  return (
    <>
      <div className="flex flex-col gap-5 h-full">
        <AppHeader title="Bitrefill" description="Live on bitcoin" />
        {open && "test"}
        <iframe
          width="100%"
          className="grow rounded-xl"
          src={`https://embed.bitrefill.com/?utm_source=bitrefill_demo&paymentMethods[]=lightning&showPaymentInfo=false`}
          sandbox="allow-same-origin allow-popups allow-scripts allow-forms"
        ></iframe>
      </div>
      <AlertDialog open={open} onOpenChange={setOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm payment</AlertDialogTitle>
            <AlertDialogDescription>
              Bitrefill asks you to pay an invoice:
              {invoice}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction>Continue</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
