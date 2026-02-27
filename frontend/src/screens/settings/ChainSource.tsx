import { AlertTriangle, Save } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

type ChainSourceType = "esplora" | "electrum" | "bitcoind_rpc" | "default";

interface FormData {
  chainSource: ChainSourceType;
  url: string;
  host: string;
  port: string;
  user: string;
  pass: string;
}

function ChainSource() {
  const { data: info } = useInfo();
  const [isLoading, setIsLoading] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const [formData, setFormData] = useState<FormData>({
    chainSource: "default",
    url: "",
    host: "",
    port: "",
    user: "",
    pass: "",
  });

  if (!info) {
    return <Loading />;
  }

  const validateForm = (): boolean => {
    setValidationError(null);
    const url = formData.url.trim();

    // Common check
    if (formData.chainSource === "esplora") {
      if (!url.startsWith("http://") && !url.startsWith("https://")) {
        setValidationError("Esplora URL must start with http:// or https://");
        return false;
      }
    }

    if (formData.chainSource === "electrum") {
      if (!url.startsWith("ssl://") && !url.startsWith("tcp://")) {
        setValidationError("Electrum URL must start with ssl:// or tcp://");
        return false;
      }
      if (!url.includes(":")) {
        setValidationError("Electrum URL must include a port (e.g. :50002).");
        return false;
      }
    }

    if (formData.chainSource === "bitcoind_rpc") {
      if (
        !formData.host ||
        !formData.port ||
        !formData.user ||
        !formData.pass
      ) {
        setValidationError(
          "All Bitcoind RPC fields (Host, Port, User, Pass) are required."
        );
        return false;
      }
      const port = Number(formData.port);
      if (!Number.isInteger(port) || port < 1 || port > 65535) {
        setValidationError("Port must be a valid number between 1 and 65535.");
        return false;
      }
    }
    return true;
  };

  const handleSaveClick = () => {
    if (validateForm()) {
      setIsDialogOpen(true);
    }
  };

  const handleSubmit = async () => {
    setIsDialogOpen(false);
    setIsLoading(true);

    try {
      const payload = {
        chainSource: formData.chainSource,
        // Only include relevant fields based on selection
        ...(formData.chainSource === "esplora" ||
        formData.chainSource === "electrum"
          ? { url: formData.url.trim() }
          : {}),
        ...(formData.chainSource === "bitcoind_rpc"
          ? {
              host: formData.host.trim(),
              port: formData.port,
              user: formData.user.trim(),
              pass: formData.pass.trim(),
            }
          : {}),
      };

      await request("/api/ldk-onchain-source", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      toast("Settings saved. Please restart your Alby Hub to apply changes.");
    } catch (error) {
      await handleRequestError("Failed to save configuration", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-2xl">
      <SettingsHeader
        title="Chain Source"
        description="Configure the onchain host to verify blockchain data from."
      />

      {/* TODO: Show the current setting 
        Possible once https://github.com/getAlby/hub/pull/2013 is merged.
        Then we can do "info?.chainDataSourceType" to pre-fill the form.
      */}

      <Alert variant="destructive" className="mt-6">
        <AlertTriangle className="h-4 w-4" />
        <AlertTitle>Warning: Advanced Setting</AlertTitle>
        <AlertDescription>
          Changing this incorrectly will prevent your node from syncing or
          starting. Ensure the source matches your network (Bitcoin/Testnet).
        </AlertDescription>
      </Alert>

      <div className="w-full mt-8 flex flex-col gap-6">
        <div className="grid gap-2">
          <Label>Source Type</Label>
          <Select
            value={formData.chainSource}
            onValueChange={(value: ChainSourceType) => {
              setFormData((prev) => ({ ...prev, chainSource: value }));
              setValidationError(null);
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select Source" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="default">Default</SelectItem>
              <SelectItem value="esplora">Esplora (HTTP)</SelectItem>
              <SelectItem value="electrum">Electrum (TCP/SSL)</SelectItem>
              <SelectItem value="bitcoind_rpc">Bitcoind RPC</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/*  ESPLORA & ELECTRUM (Single URL Field) */}
        {(formData.chainSource === "esplora" ||
          formData.chainSource === "electrum") && (
          <div className="grid gap-2">
            <Label htmlFor="url">
              {formData.chainSource === "esplora"
                ? "Esplora API URL"
                : "Electrum Server URL"}
            </Label>
            <Input
              id="url"
              type="text"
              placeholder={
                formData.chainSource === "esplora"
                  ? "https://mempool.space/api"
                  : "ssl://electrum.blockstream.info:50002"
              }
              value={formData.url}
              onChange={(e) =>
                setFormData({ ...formData, url: e.target.value })
              }
            />
            <p className="text-muted-foreground text-xs">
              {formData.chainSource === "esplora"
                ? "Must end in /api (e.g. https://mempool.space/api)"
                : "Must start with ssl:// or tcp://"}
            </p>
          </div>
        )}

        {/* BITCOIND RPC (4 Fields) */}
        {formData.chainSource === "bitcoind_rpc" && (
          <div className="grid gap-4 p-4 border rounded-md bg-muted/20">
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="host">Host</Label>
                <Input
                  id="host"
                  placeholder="127.0.0.1"
                  value={formData.host}
                  onChange={(e) =>
                    setFormData({ ...formData, host: e.target.value })
                  }
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="port">Port</Label>
                <Input
                  id="port"
                  placeholder="8332"
                  value={formData.port}
                  onChange={(e) =>
                    setFormData({ ...formData, port: e.target.value })
                  }
                />
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="user">RPC User</Label>
              <Input
                id="user"
                placeholder="bitcoin"
                value={formData.user}
                onChange={(e) =>
                  setFormData({ ...formData, user: e.target.value })
                }
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="pass">RPC Password</Label>
              <Input
                id="pass"
                type="password"
                placeholder="••••••••"
                value={formData.pass}
                onChange={(e) =>
                  setFormData({ ...formData, pass: e.target.value })
                }
              />
            </div>
          </div>
        )}

        {validationError && (
          <p className="text-destructive text-sm font-medium">
            {validationError}
          </p>
        )}

        <LoadingButton
          className="w-fit"
          loading={isLoading}
          onClick={handleSaveClick}
        >
          <Save className="w-4 h-4 mr-2" />
          Save Configuration
        </LoadingButton>
        <AlertDialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Change Chain Source?</AlertDialogTitle>
              <AlertDialogDescription>
                Your node will need to restart to apply these changes. If the
                new source is invalid, your node may fail to start.
                <br />
                <br />
                Are you sure you want to proceed?
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={handleSubmit}>
                Confirm
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  );
}

export default ChainSource;
