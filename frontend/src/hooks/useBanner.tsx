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
  const isDismissedRef = React.useRef(false);

  React.useEffect(() => {
    if (!info || !albyInfo || info.hideUpdateBanner || isDismissedRef.current) {
      return;
    }

    // vss migration (alby cloud only)
    // TODO: remove after 2026-08-01
    const vssMigrationRequired =
      info.oauthRedirect &&
      !!albyMe?.subscription.plan_code.includes("buzz") &&
      info.vssSupported &&
      !info.ldkVssEnabled;

    // Normalize both versions: info.version is usually "vX.Y.Z" (empty for
    // source builds), latestVersion comes from getalby.com and may carry a
    // leading "v" or whitespace. Never throw: an unparseable version must
    // not nag with a false-positive update banner (#1870).
    // Empty latest (unknown release) means up-to-date; empty current
    // (source build) stays not-up-to-date.
    let upToDate: boolean;
    try {
      const current = info.version?.trim().replace(/^v/i, "");
      const latest = albyInfo.hub.latestVersion?.trim().replace(/^v/i, "");
      upToDate =
        Boolean(current) && (!latest || compare(current, latest, ">="));
    } catch {
      upToDate = true;
    }

    setShowBanner(!upToDate || vssMigrationRequired);
  }, [info, albyInfo, albyMe?.subscription.plan_code]);

  const dismissBanner = () => {
    isDismissedRef.current = true;
    setShowBanner(false);
  };

  return { showBanner, dismissBanner };
}
