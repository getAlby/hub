import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { useAlbyMe } from "src/hooks/useAlbyMe";

import { useInfo } from "src/hooks/useInfo";

const PLANS: Record<string, string> = {
  buzz_202406_month: "Pro Cloud",
  buzz_202406_year: "Pro Cloud",
  buzz_202411_usd_month: "Pro Cloud",
  buzz_202411_usd_year: "Pro Cloud",
  pro_202411_usd_month: "Pro",
  pro_202411_usd_year: "Pro",
};

export function About() {
  const { data: info } = useInfo();
  const { data: albyMe, error: albyMeError } = useAlbyMe();

  if (!info || (info.albyAccountConnected && !albyMe && !albyMeError)) {
    return <Loading />;
  }

  return (
    <>
      <SettingsHeader title="About" description="Info about your Alby Hub" />
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
        <div className="grid gap-2">
          <p className="font-medium text-sm">Nostr Relay</p>
          <p className="text-muted-foreground text-sm slashed-zero">
            {info.relay}
          </p>
        </div>
        {info.albyAccountConnected && albyMe && (
          <div className="grid gap-2">
            <p className="font-medium text-sm">Connected Alby Account</p>
            <p className="text-muted-foreground text-sm slashed-zero">
              {albyMe.email || albyMe.lightning_address}
            </p>
          </div>
        )}
        <div className="grid gap-2">
          <p className="font-medium text-sm">Alby Account Plan</p>
          <p className="text-muted-foreground text-sm slashed-zero">
            {albyMe?.subscription.plan_code
              ? PLANS[albyMe.subscription.plan_code]
              : "Free"}
          </p>
        </div>
      </div>
    </>
  );
}
