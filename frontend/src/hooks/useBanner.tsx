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
    if (!info || !albyInfo || isDismissedRef.current) {
      return;
    }

    const upToDate =
      Boolean(info.version) &&
      info.version.startsWith("v") &&
      compare(info.version.substring(1), albyInfo.hub.latestVersion, ">=");

    setShowBanner(!info.hideUpdateBanner && !upToDate);
  }, [info, albyInfo, albyMe?.subscription.plan_code]);

  const dismissBanner = () => {
    isDismissedRef.current = true;
    setShowBanner(false);
  };

  return { showBanner, dismissBanner };
}
