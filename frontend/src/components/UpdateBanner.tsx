import { compare } from "compare-versions";
import { XIcon } from "lucide-react";
import React from "react";

import ExternalLink from "src/components/ExternalLink";
import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useInfo } from "src/hooks/useInfo";

export function UpdateBanner() {
  const { data: info } = useInfo();
  const { data: albyInfo } = useAlbyInfo();
  const [isVisible, setIsVisible] = React.useState(false);

  React.useEffect(() => {
    if (!info || !albyInfo) {
      setIsVisible(false);
      return;
    }

    const upToDate =
      Boolean(info.version) &&
      info.version.startsWith("v") &&
      compare(info.version.substring(1), albyInfo.hub.latestVersion, ">=");

    setIsVisible(!upToDate);
  }, [info, albyInfo]);

  const dismissBanner = () => {
    setIsVisible(false);
  };

  if (!info || !albyInfo || !isVisible) {
    return null;
  }

  return (
    <div className="fixed w-full bg-foreground text-background z-20 px-8 md:px-12 py-2 text-sm md:text-center flex items-center justify-center">
      <ExternalLink
        to={`https://getalby.com/update/hub?version=${info?.version}`}
      >
        <span className="font-semibold">Update Available</span>
        {" • "}
        <span>{albyInfo.hub.latestReleaseNotes.substring(0, 120) + "..."}</span>
      </ExternalLink>
      <XIcon
        className="absolute right-4 cursor-pointer w-4 text-background"
        onClick={dismissBanner}
      />
    </div>
  );
}
