import { useEffect, useCallback, useState } from "react";
import { toast } from "sonner";
import ModelSelect from "src/components/connections/routstr/ModelSelect";
import { RoutstrConnectionDetails } from "src/components/connections/routstr/RoutstrConnectionDetails";
import {
  ROUTSTR_AUTO_ROUTING_LABEL,
  ROUTSTR_UNIVERSAL_KEY_COPY,
} from "src/components/connections/routstr/constants";
import { Button } from "src/components/ui/button";
import { RefreshCw } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import {
  RoutstrdModel,
  getAllRoutstrdModels,
  fundFromHub,
  createRoutstrdClient,
} from "src/hooks/useRoutstrd";
import { handleRequestError } from "src/utils/handleRequestError";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onKeyCreated: (apiKey: string, clientId?: string) => void;
  connectionAppId: number;
};

export function CreateKeyDialog({
  open,
  onOpenChange,
  onKeyCreated,
  connectionAppId,
}: Props) {
  const [step, setStep] = useState<"setup" | "paying" | "done">("setup");
  const [models, setModels] = useState<RoutstrdModel[]>([]);
  const [modelsLoading, setModelsLoading] = useState(false);
  const [amount, setAmount] = useState("5000");
  const [isProcessing, setIsProcessing] = useState(false);
  const [createdKey, setCreatedKey] = useState("");

  const loadModels = useCallback(async (refresh = false) => {
    setModelsLoading(true);
    try {
      const result = await getAllRoutstrdModels(refresh);
      if (result?.models) {
        setModels(result.models);
      }
    } catch (e) {
      toast.error("Could not load models. Is routstrd running?");
      console.error(e);
    } finally {
      setModelsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      setStep("setup");
      setAmount("5000");
      setCreatedKey("");
      setIsProcessing(false);
      loadModels();
    }
  }, [open, loadModels]);

  const handleCreateKey = async () => {
    setIsProcessing(true);
    try {
      // 1. Create the API key first (free). Funding before key creation
      // would leave sats spent with no key if the daemon call failed.
      const clientName = `routstr-app-${connectionAppId}-${Date.now()}`;
      const clientResult = await createRoutstrdClient(clientName);
      const apiKey = clientResult?.client?.apiKey;
      const clientId = clientResult?.client?.id || clientName;
      if (!apiKey) {
        throw new Error("Failed to create API key");
      }

      // 2. Fund via Hub's LN from this app's wallet
      if (Number(amount) > 0) {
        setStep("paying");
        await fundFromHub(Number(amount), connectionAppId);
      }

      setCreatedKey(apiKey);
      setStep("done");
      onKeyCreated(apiKey, clientId);
      toast.success("Deposit & key created!");
    } catch (error) {
      handleRequestError("Failed to create API key", error);
      setStep("setup");
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[520px]">
        {step === "setup" && (
          <>
            <DialogHeader>
              <DialogTitle>Create API Key</DialogTitle>
              <DialogDescription>
                Deposit sats into the Routstr wallet, then create a universal
                API key. {ROUTSTR_UNIVERSAL_KEY_COPY}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-2">
              <div>
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Models available</Label>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {ROUTSTR_AUTO_ROUTING_LABEL}
                    </p>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => loadModels(true)}
                    disabled={modelsLoading}
                  >
                    <RefreshCw
                      className={`h-3.5 w-3.5 ${modelsLoading ? "animate-spin" : ""}`}
                    />
                  </Button>
                </div>
                <div className="mt-2">
                  <ModelSelect
                    models={models.filter((m) => m.enabled !== false)}
                    browseOnly
                    disabled={modelsLoading}
                  />
                </div>
              </div>

              <div>
                <Label htmlFor="deposit-amount">Deposit (sats)</Label>
                <Input
                  id="deposit-amount"
                  type="number"
                  min={1}
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Funds the Routstr Cashu wallet used for inference (not held
                  inside the key string). Paid via Hub Lightning (instant).
                </p>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <LoadingButton
                loading={isProcessing}
                onClick={handleCreateKey}
                disabled={!amount || Number(amount) < 1}
              >
                Deposit & Create Key
              </LoadingButton>
            </DialogFooter>
          </>
        )}

        {step === "paying" && (
          <>
            <DialogHeader>
              <DialogTitle>Funding wallet…</DialogTitle>
              <DialogDescription>
                Depositing {amount} sats and generating your API key.
              </DialogDescription>
            </DialogHeader>
            <div className="py-8 text-center text-sm text-muted-foreground">
              Please wait…
            </div>
          </>
        )}

        {step === "done" && createdKey && (
          <>
            <DialogHeader>
              <DialogTitle>API key ready</DialogTitle>
              <DialogDescription>
                Copy these details into your OpenAI-compatible client.
              </DialogDescription>
            </DialogHeader>
            <div className="py-2 max-h-[60vh] overflow-y-auto">
              <RoutstrConnectionDetails apiKey={createdKey} showHubUrl />
            </div>
            <DialogFooter>
              <Button onClick={() => onOpenChange(false)}>Done</Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
