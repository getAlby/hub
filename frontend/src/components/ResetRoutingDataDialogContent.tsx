import React from "react";
import { toast } from "sonner";
import {
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

const RESET_KEY_OPTIONS = [
  {
    value: "ALL",
    label: "All",
    description: "Clears both the scorer, network graph data, and node metrics",
  },
  {
    value: "Scorer",
    label: "Scorer",
    description:
      "Clears the scores/penalties applied to nodes from past payment attempts.",
  },
  {
    value: "NetworkGraph",
    label: "Network Graph",
    description: "Clears the cache of nodes on the network",
  },
  {
    value: "NodeMetrics",
    label: "Node Metrics",
    description:
      "Clears last sync timestamps to do a full wallet or network graph re-scan when RGS is enabled",
  },
];

export function ResetRoutingDataDialogContent() {
  const { mutate: reloadInfo } = useInfo();
  const [resetKey, setResetKey] = React.useState<string>();

  async function resetRouter() {
    try {
      await request("/api/reset-router", {
        method: "POST",
        body: JSON.stringify({ key: resetKey }),
        headers: {
          "Content-Type": "application/json",
        },
      });
      await reloadInfo();
      toast("🎉 Router reset");
    } catch (error) {
      console.error(error);
      toast.error("Something went wrong", {
        description: "" + error,
      });
    }
  }

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Clear Routing Data</AlertDialogTitle>
        <AlertDialogDescription className="text-left">
          <div>
            <p>Are you sure you want to clear your routing data?</p>
            <div className="grid gap-2 mt-4">
              <Label className="text-primary">Routing Data to Clear</Label>
              <Select
                name="resetKey"
                value={resetKey}
                onValueChange={(value) => setResetKey(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select Data" />
                </SelectTrigger>
                <SelectContent>
                  {RESET_KEY_OPTIONS.map((resetKey) => (
                    <SelectItem key={resetKey.value} value={resetKey.value}>
                      {resetKey.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2 mt-4 border rounded-md p-3">
              <h3 className="text-primary font-semibold">Clear Data Options</h3>
              {RESET_KEY_OPTIONS.map((resetKey) => (
                <p>
                  <span className="text-primary font-medium">
                    {resetKey.label}
                  </span>
                  {" - "}
                  {resetKey.description}
                </p>
              ))}
            </div>
            <p className="text-primary font-medium mt-4">
              After clearing, you'll need to login again to restart your node.
            </p>
          </div>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel onClick={() => setResetKey(undefined)}>
          Cancel
        </AlertDialogCancel>
        <AlertDialogAction disabled={!resetKey} onClick={resetRouter}>
          Confirm
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}
