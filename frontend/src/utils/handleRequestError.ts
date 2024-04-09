export function handleRequestError(
  toast: any,
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
