import { useState } from "react";
import { Button } from "src/components/ui/button";
import { Textarea } from "src/components/ui/textarea";

import { request } from "src/utils/request";

export default function DebugTools() {
  const [apiResponse, setApiResponse] = useState<string>("");

  async function apiRequest(
    endpoint: string,
    method: string,
    requestBody?: object
  ) {
    try {
      const requestOptions: RequestInit = {
        method: method,
        headers: {
          "Content-Type": "application/json",
        },
      };

      if (requestBody) {
        requestOptions.body = JSON.stringify(requestBody);
      }

      const data = await request(endpoint, requestOptions);

      setApiResponse(
        (data as { logs: string }).logs || JSON.stringify(data, null, 2)
      );
    } catch (error) {
      setApiResponse(JSON.stringify(error, Object.getOwnPropertyNames(error)));
    }
  }

  return (
    <div>
      <div className="grid gap-6 m-8 md:grid-cols-2 xl:grid-cols-3">
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
            const amount = window.prompt("Enter amount in sats:");
            if (amount) {
              const nodeId = window.prompt("Enter node pubkey:");
              if (nodeId) {
                apiRequest("/api/send-spontaneous-payment-probes", "POST", {
                  amount: parseInt(amount) * 1000,
                  nodeId,
                });
              }
            }
          }}
        >
          Probe Keysend
        </Button>
        <Button onClick={() => apiRequest("/api/peers", "GET")}>
          List Peers
        </Button>
        <Button onClick={() => apiRequest("/api/channels", "GET")}>
          List Channels
        </Button>
        <Button
          onClick={() => {
            const maxLen = window.prompt("Enter max length (in characters):");

            if (maxLen) {
              apiRequest(`/api/log/app?maxLen=${maxLen}`, "GET");
            }
          }}
        >
          Get App Logs
        </Button>
        <Button
          onClick={() => {
            const maxLen = window.prompt("Enter max length (in characters):");

            if (maxLen) {
              apiRequest(`/api/log/node?maxLen=${maxLen}`, "GET");
            }
          }}
        >
          Get Node Logs
        </Button>
        <Button
          onClick={() => {
            apiRequest(`/api/node/status`, "GET");
          }}
        >
          Get Node Status
        </Button>
        <Button
          onClick={() => {
            apiRequest(`/api/balances`, "GET");
          }}
        >
          Get Balances
        </Button>
        <Button
          onClick={() => {
            const nodeIds = window.prompt(
              "Enter node IDs, comma-separated:",
              "030a58b8653d32b99200a2334cfe913e51dc7d155aa0116c176657a4f1722677a3,03b65c19de2da9d35895b37e6fa263cefbf8d184e6b61c0713747ebf12409d219f"
            );
            if (nodeIds) {
              apiRequest(`/api/node/network-graph?nodeIds=${nodeIds}`, "GET");
            }
          }}
        >
          Get Network Graph
        </Button>
      </div>
      {apiResponse && (
        <Textarea
          className="whitespace-pre-wrap break-words font-emoji"
          rows={35}
          value={`API Response: ${apiResponse}`}
        />
      )}
    </div>
  );
}
