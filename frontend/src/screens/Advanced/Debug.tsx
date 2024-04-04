import { useState } from "react";
import { Button } from "src/components/ui/button";
import { Card, CardContent } from "src/components/ui/card";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";

export default function Debug() {
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
          Send Payment Probes
        </Button>
        <Button
          onClick={() => {
            const amount_msat = window.prompt("Enter amount in milli satoshi:");
            const node_id = window.prompt("Enter node_id:");
            if (amount_msat && node_id)
              apiRequest("/api/send-spontaneous-payment-probes", "POST", {
                amount_msat: parseInt(amount_msat),
                node_id: node_id,
              });
          }}
        >
          Send Spontaneous Payment Probes
        </Button>
        <Button onClick={() => apiRequest("/api/peers", "GET")}>
          List Peers
        </Button>
        <Button
          onClick={() => {
            let maxLen = window.prompt("Enter max length:");

            if (maxLen)
              apiRequest("/api/get-log-output", "POST", {
                maxLen: parseInt(maxLen),
              });
          }}
        >
          Get Log Output
        </Button>
      </div>
      {apiResponse && (
        <Card className="mt-8 pt-6">
          <CardContent>
            <pre>API Response: {JSON.stringify(apiResponse, null, 2)}</pre>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
