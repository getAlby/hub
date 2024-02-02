import toast from "src/components/Toast";

export function handleRequestError(message: string, error: unknown) {
  console.error(message, error);
  toast.error(message + ": " + error);
}
