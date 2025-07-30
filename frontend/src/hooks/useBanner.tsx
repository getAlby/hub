import { compare } from "compare-versions";
import React from "react";
import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export function useBanner() {
  const { data: info } = useInfo();
  const { data: albyInfo } = useAlbyInfo();
  const { data: albyMe } = useAlbyMe();
  const [showBanner, setShowBanner] = React.useState(false);

  React.useEffect(() => {
    if (!info || !albyInfo) {
      setShowBanner(false);
      return;
    }

    // vss migration (alby cloud only)
    // TODO: remove after 2026-01-01
    const vssMigrationRequired =
      info.oauthRedirect &&
      !!albyMe?.subscription.plan_code.includes("buzz") &&
      info.vssSupported &&
      !info.ldkVssEnabled;

    const upToDate =
      Boolean(info.version) &&
      info.version.startsWith("v") &&
      compare(info.version.substring(1), albyInfo.hub.latestVersion, ">=");

    setShowBanner(!upToDate || vssMigrationRequired);
  }, [info, albyInfo, albyMe?.subscription.plan_code]);

  const dismissBanner = () => {
    setShowBanner(false);
  };

  return { showBanner, dismissBanner };
}
