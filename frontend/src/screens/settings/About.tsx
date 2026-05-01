import { ExternalLinkIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Badge } from "src/components/ui/badge";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useNodeDetails } from "src/hooks/useNodeDetails";

import { useInfo } from "src/hooks/useInfo";

export function About() {
  const { data: info } = useInfo();
  const { data: albyMe, error: albyMeError } = useAlbyMe();
  const lsps2Source = info?.jitChannelsLiquiditySource;
  const lsps2Pubkey = lsps2Source?.includes("@")
    ? lsps2Source.split("@")[0]
    : undefined;
  const { data: lsps2NodeDetails } = useNodeDetails(lsps2Pubkey);
  const lsps2Label =
    lsps2NodeDetails?.alias ||
    (lsps2Pubkey ? lsps2Pubkey.slice(0, 8) + "..." : lsps2Source);

  if (!info || (info.albyAccountConnected && !albyMe && !albyMeError)) {
    return <Loading />;
  }

  const planCode = albyMe?.subscription.plan_code;
  let planName: string;
  if (planCode?.startsWith("buzz_")) {
    planName = "Pro Cloud";
  } else if (planCode?.startsWith("pro_")) {
    planName = "Pro";
  } else if (planCode) {
    planName = "Paid";
  } else {
    planName = "Free";
  }

  return (
    <>
      <SettingsHeader
        pageTitle="About"
        title="About"
        description="Info about your Alby Hub"
      />
      <div className="grid gap-4">
        <div className="grid gap-2">
          <p className="font-medium text-sm">Alby Hub Version</p>
          <p className="text-muted-foreground text-sm slashed-zero">
            {info.version || "dev"}
          </p>
        </div>
        <div className="grid gap-2">
          <p className="font-medium text-sm">Lightning Node Backend</p>
          <p className="text-muted-foreground text-sm slashed-zero">
            {info.backendType}
          </p>
        </div>
        {info.chainDataSourceType && (
          <div className="grid gap-2">
            <p className="font-medium text-sm">Chain Data Source</p>
            <div className="flex flex-col gap-1 text-sm text-muted-foreground">
              <p className="capitalize">{info.chainDataSourceType}</p>
              {info.chainDataSourceAddress && (
                <p className="break-all">{info.chainDataSourceAddress}</p>
              )}
            </div>
          </div>
        )}
        {info.jitChannelsLiquiditySource && (
          <div className="grid gap-2">
            <p className="font-medium text-sm">
              Just-in-Time channels Liquidity Source LSP2
            </p>
            <div className="flex flex-col gap-1 text-sm text-muted-foreground">
              {lsps2Pubkey ? (
                <ExternalLink
                  to={`${info.mempoolUrl}/lightning/node/${lsps2Pubkey}`}
                  className="inline-flex items-center gap-1 underline w-fit"
                >
                  {lsps2Label}
                  <ExternalLinkIcon className="size-4" />
                </ExternalLink>
              ) : (
                <p>{lsps2Label}</p>
              )}
              <p className="break-all">{info.jitChannelsLiquiditySource}</p>
            </div>
          </div>
        )}
        <div className="grid gap-2">
          <p className="font-medium text-sm">Nostr Relays</p>
          {info.relays.map((relay) => (
            <p className="flex items-center gap-2 text-muted-foreground text-sm">
              {relay.url}
              <Badge variant={relay.online ? "positive" : "destructive"}>
                {relay.online ? "online" : "offline"}
              </Badge>
            </p>
          ))}
        </div>
        {info.albyAccountConnected && albyMe && (
          <div className="grid gap-2">
            <p className="font-medium text-sm">Connected Alby Account</p>
            <p className="text-muted-foreground text-sm slashed-zero">
              {albyMe.email || albyMe.lightning_address}
            </p>
          </div>
        )}
        {info.albyAccountConnected && albyMe?.hub.name && (
          <div className="grid gap-2">
            <p className="font-medium text-sm">Alby Hub Name</p>
            <p className="text-muted-foreground text-sm slashed-zero">
              {albyMe.hub.name}
            </p>
          </div>
        )}
        {info.albyAccountConnected && (
          <div className="grid gap-2">
            <p className="font-medium text-sm">Alby Account Plan</p>
            <p className="text-muted-foreground text-sm slashed-zero">
              {planName}
            </p>
          </div>
        )}
      </div>
    </>
  );
}
