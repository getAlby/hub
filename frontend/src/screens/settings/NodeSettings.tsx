import { toast } from "sonner";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Label } from "src/components/ui/label";
import { Switch } from "src/components/ui/switch";

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

  const hasJitSource = !!info.jitChannelsLiquiditySource;

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
          ? "Just-in-time channels enabled"
          : "Just-in-time channels disabled"
      );
    } catch (error) {
      handleRequestError("Failed to update just-in-time channels", error);
    }
  }

  return (
    <>
      <SettingsHeader
        pageTitle="Node"
        title="Node"
        description="Configure how your node handles channels and payments."
      />
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between gap-8">
          <div className="flex flex-col gap-1">
            <Label htmlFor="jit-channels" className="cursor-pointer">
              Just-in-time channels
            </Label>
            <p className="text-sm text-muted-foreground">
              Automatically open a new channel through a liquidity provider when
              you receive a payment that exceeds your receive limit. The
              provider's fee is deducted from that payment.{" "}
              <ExternalLink
                to="https://guides.getalby.com/user-guide/alby-hub/faq/what-are-just-in-time-channels"
                className="underline"
              >
                Learn more
              </ExternalLink>
            </p>
          </div>
          <Switch
            id="jit-channels"
            checked={info.jitChannelsEnabled}
            disabled={!hasJitSource}
            onCheckedChange={setJitChannelsEnabled}
          />
        </div>
        {!hasJitSource && (
          <p className="text-sm text-muted-foreground">
            Just-in-time channels are currently not available on your network.
          </p>
        )}
      </div>
    </>
  );
}
