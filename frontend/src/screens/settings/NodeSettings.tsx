import { InfoIcon } from "lucide-react";

import { toast } from "sonner";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";

import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function NodeSettings() {
  const { data: info, mutate: refetchInfo } = useInfo();

  if (!info) {
    return <Loading />;
  }
  if (info.backendType !== "LDK") {
    return <p>Your Hub does not support this feature.</p>;
  }

  async function setJitChannelsEnabled(enabled: boolean) {
    try {
      await request("/api/settings", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ jitChannelsEnabled: enabled }),
      });
      await refetchInfo();
      toast(
        enabled
          ? "JIT channels enabled. Restart your Hub to apply."
          : "JIT channels disabled. Restart your Hub to apply."
      );
    } catch (error) {
      handleRequestError("Failed to update JIT channels setting", error);
    }
  }

  return (
    <>
      <SettingsHeader
        pageTitle="Node"
        title="Node"
        description="Configure your node's behavior"
      />
      <div className="flex flex-col gap-4">
        <div>
          <p className="text-muted-foreground">
            JIT (just-in-time) channels let you receive payments larger than
            your current inbound capacity by automatically opening a new channel
            through a liquidity provider. The provider's fee is deducted from
            the incoming payment.
          </p>
        </div>
        <div className="flex items-center">
          <Checkbox
            id="jit-channels"
            checked={info.jitChannelsEnabled}
            onCheckedChange={(checked) =>
              setJitChannelsEnabled(checked === true)
            }
          />
          <Label htmlFor="jit-channels" className="ml-2 cursor-pointer">
            Enable JIT channels for receiving
          </Label>
        </div>
        <Alert>
          <InfoIcon />
          <AlertTitle>Restart required</AlertTitle>
          <AlertDescription>
            Changes to this setting only take effect after you restart your Hub.
          </AlertDescription>
        </Alert>
      </div>
    </>
  );
}
