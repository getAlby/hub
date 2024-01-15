import toast from "src/components/Toast";

export async function validateFetchResponse(response: Response) {
  if (!response.ok) {
    let reason = "unknown";
    try {
      reason = await response.text();
    } catch (error) {
      console.error("Failed to read response text", error);
    }
    throw new Error("Unexpected response: " + response.status + " " + reason);
  }
}

export function handleFetchError(message: string, error: unknown) {
  console.error(message, error);
  toast.error(message + ": " + error);
}
