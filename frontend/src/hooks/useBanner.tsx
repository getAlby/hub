import { compare } from "compare-versions";
import React from "react";
import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useInfo } from "src/hooks/useInfo";

export function useBanner() {
  const { data: info } = useInfo();
  const { data: albyInfo } = useAlbyInfo();
  const [showBanner, setShowBanner] = React.useState(false);

  React.useEffect(() => {
    if (!info || !albyInfo) {
      setShowBanner(false);
      return;
    }

    const upToDate =
      Boolean(info.version) &&
      info.version.startsWith("v") &&
      compare(info.version.substring(1), albyInfo.hub.latestVersion, ">=");

    setShowBanner(!upToDate);
  }, [info, albyInfo]);

  const dismissBanner = () => {
    setShowBanner(false);
  };

  return { showBanner, dismissBanner };
}
