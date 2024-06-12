import { Toast, ToasterToast } from "src/components/ui/use-toast";

type ToastSignature = (props: Toast) => {
  id: string;
  dismiss: () => void;
  update: (props: ToasterToast) => void;
};

export function handleRequestError(
  toast: ToastSignature,
  message: string,
  error: unknown
) {
  console.error(message, error);
  toast({
    title: message,
    description: isErrorWithMessage(error) ? error.message : null,
    variant: "destructive",
  });
}
type ErrorWithMessage = {
  message: string;
};

function isErrorWithMessage(error: unknown): error is ErrorWithMessage {
  return (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof (error as Record<string, unknown>).message === "string"
  );
}
