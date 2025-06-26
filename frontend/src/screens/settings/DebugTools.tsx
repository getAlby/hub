import React from "react";
import { ExecuteCustomNodeCommandDialogContent } from "src/components/ExecuteCustomNodeCommandDialogContent";
import { ResetRoutingDataDialogContent } from "src/components/ResetRoutingDataDialogContent";
import SettingsHeader from "src/components/SettingsHeader";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { Textarea } from "src/components/ui/textarea";
import { useInfo } from "src/hooks/useInfo";

import { request } from "src/utils/request";

type Props = {
  apiRequest: (
    endpoint: string,
    method: string,
    requestBody?: object
  ) => Promise<void>;
  target?: string;
};

function ProbeInvoiceDialogContent({ apiRequest }: Props) {
  const [invoice, setInvoice] = React.useState<string>();

  async function onConfirm() {
    await apiRequest("/api/send-payment-probes", "POST", {
      invoice: invoice,
    });
    setInvoice("");
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Probe Invoice</AlertDialogTitle>
        <AlertDialogDescription className="text-start">
          <Label htmlFor="invoice" className="block mb-2">
            Enter Invoice
          </Label>
          <Input
            id="invoice"
            name="invoice"
            type="text"
            placeholder="lnbc...."
            required
            autoFocus
            value={invoice}
            onChange={(e) => {
              setInvoice(e.target.value.trim());
            }}
          />
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction disabled={!invoice} onClick={onConfirm}>
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}

function ProbeKeysendDialogContent({ apiRequest }: Props) {
  const [amount, setAmount] = React.useState<string>("");
  const [nodeId, setNodeId] = React.useState<string>("");

  async function onConfirm() {
    await apiRequest("/api/send-spontaneous-payment-probes", "POST", {
      amount: (parseInt(amount) || 0) * 1000,
      nodeId,
    });
    setAmount("");
    setNodeId("");
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Probe Keysend</AlertDialogTitle>
        <AlertDialogDescription className="text-start">
          <div>
            <Label htmlFor="amount" className="block mb-2">
              Enter Amount (sats)
            </Label>
            <Input
              id="amount"
              name="amount"
              type="number"
              min={0}
              autoFocus
              value={amount}
              onChange={(e) => {
                setAmount(e.target.value.trim());
              }}
            />
          </div>
          <div className="mt-4">
            <Label htmlFor="pubkey" className="block mb-2">
              Enter Node Pubkey
            </Label>
            <Input
              id="pubkey"
              type="text"
              value={nodeId}
              onChange={(e) => {
                setNodeId(e.target.value.trim());
              }}
            />
          </div>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction
          disabled={!parseInt(amount) || !nodeId}
          onClick={onConfirm}
        >
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}

function GetLogsDialogContent({ apiRequest, target }: Props) {
  const [maxLen, setMaxLen] = React.useState<string>("");

  async function onConfirm() {
    await apiRequest(`/api/log/${target}?maxLen=${maxLen}`, "GET");
    setMaxLen("");
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle className="capitalize">
          Get {target} Logs
        </AlertDialogTitle>
        <AlertDialogDescription className="text-start">
          <Label htmlFor="maxLength" className="block mb-2">
            Enter Max Length (in characters)
          </Label>
          <Input
            id="maxLength"
            name="maxLength"
            type="number"
            required
            autoFocus
            min={1}
            value={maxLen}
            onChange={(e) => {
              setMaxLen(e.target.value.trim());
            }}
          />
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction disabled={!parseInt(maxLen)} onClick={onConfirm}>
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}

function GetNetworkGraphDialogContent({ apiRequest }: Props) {
  const [nodeIds, setNodeIds] = React.useState<string>("");

  async function onConfirm() {
    await apiRequest(`/api/node/network-graph?nodeIds=${nodeIds}`, "GET");
    setNodeIds("");
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Get Network Graph</AlertDialogTitle>
        <AlertDialogDescription className="text-start">
          <Label htmlFor="nodes" className="block mb-2">
            Enter Node Pubkeys (separated by commas)
          </Label>
          <Input
            id="nodes"
            type="text"
            placeholder="e.g. nodepubkey1,nodepubkey2,nodepubkey3"
            value={nodeIds}
            onChange={(e) => {
              setNodeIds(e.target.value.trim());
            }}
          />
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction disabled={!nodeIds} onClick={onConfirm}>
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}

export default function DebugTools() {
  const [apiResponse, setApiResponse] = React.useState<string>("");
  const [dialog, setDialog] = React.useState<
    | "probeInvoice"
    | "probeKeysend"
    | "getAppLogs"
    | "getNodeLogs"
    | "getNetworkGraph"
    | "resetRoutingData"
    | "customNodeCommand"
  >();

  const { data: info } = useInfo();

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
      <SettingsHeader
        title="Debug Tools"
        description="Extra tools for debugging purposes."
      />
      <div className="grid mt-6 gap-6 mb-8 lg:mb-8 md:grid-cols-2 xl:grid-cols-3">
        <AlertDialog
          onOpenChange={() => {
            if (!open) {
              setDialog(undefined);
            }
          }}
        >
          <Button
            variant={"outline"}
            onClick={() => apiRequest("/api/info", "GET")}
          >
            Get Info
          </Button>
          <Button
            variant={"outline"}
            onClick={() => apiRequest("/api/peers", "GET")}
          >
            List Peers
          </Button>
          <Button
            variant={"outline"}
            onClick={() => apiRequest("/api/channels", "GET")}
          >
            List Channels
          </Button>
          <AlertDialogTrigger asChild>
            <Button variant={"outline"} onClick={() => setDialog("getAppLogs")}>
              Get App Logs
            </Button>
          </AlertDialogTrigger>
          <AlertDialogTrigger asChild>
            <Button
              variant={"outline"}
              onClick={() => setDialog("getNodeLogs")}
            >
              Get Node Logs
            </Button>
          </AlertDialogTrigger>
          <Button
            variant={"outline"}
            onClick={() => {
              apiRequest(`/api/node/status`, "GET");
            }}
          >
            Get Node Status
          </Button>
          <Button
            variant={"outline"}
            onClick={() => {
              apiRequest(`/api/balances`, "GET");
            }}
          >
            Get Balances
          </Button>
          <AlertDialogTrigger asChild>
            <Button
              variant={"outline"}
              onClick={() => setDialog("getNetworkGraph")}
            >
              Get Network Graph
            </Button>
          </AlertDialogTrigger>
          {(info?.backendType === "LDK" || info?.backendType === "CASHU") && (
            <AlertDialogTrigger asChild>
              <Button
                variant={"outline"}
                onClick={() => setDialog("resetRoutingData")}
              >
                Clear Routing Data
              </Button>
            </AlertDialogTrigger>
          )}
          <Button
            variant={"outline"}
            onClick={() => {
              apiRequest(`/api/commands`, "GET");
            }}
          >
            Get Node Commands
          </Button>
          <AlertDialogTrigger asChild>
            <Button
              variant={"outline"}
              onClick={() => {
                apiRequest(`/api/commands`, "GET");
                setDialog("customNodeCommand");
              }}
            >
              Execute Node Command
            </Button>
          </AlertDialogTrigger>
          {info?.backendType === "LDK" && (
            <Button
              variant={"outline"}
              onClick={() => {
                apiRequest(`/api/command`, "POST", {
                  command: "export_pathfinding_scores",
                });
              }}
            >
              Export Pathfinding Scores
            </Button>
          )}
          {/* probing functions are not useful */}
          {/*info?.backendType === "LDK" && (
            <AlertDialogTrigger asChild>
              <Button onClick={() => setDialog("probeInvoice")}>
                Probe Invoice
              </Button>
            </AlertDialogTrigger>
          )*/}
          {/*info?.backendType === "LDK" && (
            <AlertDialogTrigger asChild>
              <Button onClick={() => setDialog("probeKeysend")}>
                Probe Keysend
              </Button>
            </AlertDialogTrigger>
          )*/}

          {dialog === "probeInvoice" && (
            <ProbeInvoiceDialogContent apiRequest={apiRequest} />
          )}
          {dialog === "probeKeysend" && (
            <ProbeKeysendDialogContent apiRequest={apiRequest} />
          )}
          {(dialog === "getAppLogs" || dialog === "getNodeLogs") && (
            <GetLogsDialogContent
              apiRequest={apiRequest}
              target={dialog === "getAppLogs" ? "app" : "node"}
            />
          )}
          {dialog === "getNetworkGraph" && (
            <GetNetworkGraphDialogContent apiRequest={apiRequest} />
          )}
          {dialog === "resetRoutingData" && <ResetRoutingDataDialogContent />}
          {dialog === "customNodeCommand" && (
            <ExecuteCustomNodeCommandDialogContent
              availableCommands={apiResponse}
              setCommandResponse={setApiResponse}
            />
          )}
        </AlertDialog>
      </div>
      {apiResponse && (
        <Textarea
          className="whitespace-pre-wrap break-words font-mono"
          rows={35}
          value={`API Response: ${apiResponse}`}
        />
      )}
    </div>
  );
}
