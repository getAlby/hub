import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { Badge } from "src/components/ui/badge";
import { useAlbyMe } from "src/hooks/useAlbyMe";

import { useInfo } from "src/hooks/useInfo";

export function About() {
  const { data: info } = useInfo();
  const { data: albyMe, error: albyMeError } = useAlbyMe();

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
