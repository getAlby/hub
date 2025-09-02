import { toast } from "sonner";

export function handleRequestError(message: string, error: unknown) {
  console.error(message, error);
  toast.error(message, {
    description: isErrorWithMessage(error) ? error.message : undefined,
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
