import { CreateAppRequest, CreateAppResponse } from "src/types";
import { request } from "src/utils/request";

export async function createApp(
  createAppRequest: CreateAppRequest
): Promise<CreateAppResponse> {
  const createAppResponse = await request<CreateAppResponse>("/api/apps", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(createAppRequest),
  });

  if (!createAppResponse) {
    throw new Error("no create app response received");
  }
  return createAppResponse;
}
