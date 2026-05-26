import { request } from "src/utils/request";

export async function sendEvent(
  name: string,
  properties?: Record<string, unknown>
) {
  try {
    await request(`/api/event`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ event: name, properties }),
    });
  } catch (error) {
    console.error(error);
  }
}
