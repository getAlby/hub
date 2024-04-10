import { useState } from "react";
import { Button } from "src/components/ui/button";
import { Textarea } from "src/components/ui/textarea";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";

export default function DebugTools() {
  const { data: csrf } = useCSRF();
  const [apiResponse, setApiResponse] = useState<any>(null);

  async function apiRequest(
    endpoint: string,
    method: string,
    requestBody?: object
  ) {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }

      const requestOptions: RequestInit = {
        method: method,
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
      };

      if (requestBody) {
        requestOptions.body = JSON.stringify(requestBody);
      }

      const data = await request(endpoint, requestOptions);

      setApiResponse(data);
    } catch (error) {
      setApiResponse(error);
    }
  }

  return (
    <div>
      <div className="grid gap-6 m-8 md:grid-cols-3 xl:grid-cols-4">
        <Button
          onClick={() => {
            const invoice = window.prompt("enter invoice");
            apiRequest("/api/send-payment-probes", "POST", {
              invoice: invoice,
            });
          }}
        >
          Probe Invoice
        </Button>
        <Button
          onClick={() => {
            const amount_msat = window.prompt("Enter amount (millisatoshi):");
            const node_id = window.prompt("Enter node_id:");
            if (amount_msat && node_id)
              apiRequest("/api/send-spontaneous-payment-probes", "POST", {
                amount: parseInt(amount_msat) * 1000,
                nodeID: node_id,
              });
          }}
        >
          Probe Keysend
        </Button>
        <Button onClick={() => apiRequest("/api/peers", "GET")}>
          List Peers
        </Button>
        <Button
          onClick={() => {
            const maxLen = window.prompt("Enter max length (in characters):");

            if (maxLen) apiRequest(`/api/log/app?maxLen=${maxLen}`, "GET");
          }}
        >
          Get App Logs
        </Button>
        <Button
          onClick={() => {
            const maxLen = window.prompt("Enter max length (in characters):");

            if (maxLen) apiRequest(`/api/log/node?maxLen=${maxLen}`, "GET");
          }}
        >
          Get Node Logs
        </Button>
      </div>
      {apiResponse && (
        <Textarea
          className="whitespace-pre-wrap break-all"
          rows={35}
          placeholder={`API Response: ${JSON.stringify(apiResponse, null, 2)}`}
        />
      )}
    </div>
  );
}
