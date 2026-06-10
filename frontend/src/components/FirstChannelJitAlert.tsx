import { InfoIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";

export default function FirstChannelJitAlert() {
  const { data: info } = useInfo();
  const { data: channels } = useChannels();

  // a JIT channel only opens when the feature is enabled AND an LSPS2 liquidity
  // source is actually configured (jitChannelsEnabled alone is just a settings
  // toggle and can be true on backends without an LSPS2 source).
  const lsps2Source = info?.jitChannelsEnabled
    ? info.jitChannelsLiquiditySource
    : undefined;

  // only relevant when the user has no channels yet - their first received
  // payment will open the channel.
  if (!lsps2Source || !channels || channels.length > 0) {
    return null;
  }

  const minPaymentSizeMsat = info?.jitChannelsMinPaymentSizeMsat;

  return (
    <Alert>
      <InfoIcon className="h-4 w-4" />
      <AlertTitle>First payment opens a channel</AlertTitle>
      <AlertDescription className="inline">
        A channel fee applies.{" "}
        {!!minPaymentSizeMsat && (
          <>
            Minimum payment{" "}
            <FormattedBitcoinAmount amountMsat={minPaymentSizeMsat} />.{" "}
          </>
        )}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/alby-hub/faq/what-are-just-in-time-channels"
          className="underline"
        >
          Learn more
        </ExternalLink>
      </AlertDescription>
    </Alert>
  );
}
