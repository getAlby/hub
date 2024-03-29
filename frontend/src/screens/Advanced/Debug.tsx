import { useState } from "react";
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
        <button
          onClick={() => {
            const invoice = window.prompt("enter invoice");
            apiRequest("/api/send-payment-probes", "POST", {
              invoice: invoice,
            });
          }}
          className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 mr-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800"
        >
          Send Payment Probes
        </button>
        <button
          onClick={() => {
            const amount_msat = window.prompt("Enter amount in milli satoshi:");
            const node_id = window.prompt("Enter node_id:");
            if (amount_msat && node_id)
              apiRequest("/api/send-spontaneous-payment-probes", "POST", {
                amount_msat: parseInt(amount_msat),
                node_id: node_id,
              });
          }}
          className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 mr-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800"
        >
          Send Spontaneous Payment Probes
        </button>
        <button
          onClick={() => apiRequest("/api/peers", "GET")}
          className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 mr-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800"
        >
          List Peers
        </button>
        <button
          onClick={() => {
            let maxLen = window.prompt("Enter max length:");

            if (maxLen)
              apiRequest("/api/get-log-output", "POST", {
                maxLen: parseInt(maxLen),
              });
          }}
          className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 mr-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800"
        >
          Get Log Output
        </button>
      </div>
      {apiResponse && (
        <div className="mt-8 p-2 bg-black text-white overflow-x-auto">
          <pre>API Response: {JSON.stringify(apiResponse, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}
