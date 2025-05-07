import { XIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";

import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useInfo } from "src/hooks/useInfo";

export function UpdateBanner({ onDismiss }: { onDismiss: () => void }) {
  const { data: info } = useInfo();
  const { data: albyInfo } = useAlbyInfo();

  if (!info || !albyInfo) {
    return null;
  }

  return (
    <div className="fixed w-full bg-foreground text-background z-20 py-2 text-sm flex items-center justify-center">
      <ExternalLink
        to={`https://getalby.com/update/hub?version=${info?.version}`}
        className="flex items-center max-w-[80%]"
      >
        <p className="line-clamp-2 md:block whitespace-normal md:whitespace-nowrap overflow-hidden text-ellipsis">
          <span className="font-semibold mr-2">Update Available</span>
          <span>•</span>
          <span className="ml-2">{albyInfo.hub.latestReleaseNotes}</span>
        </p>
      </ExternalLink>
      <XIcon
        className="absolute right-4 cursor-pointer w-4 text-background"
        onClick={onDismiss}
      />
    </div>
  );
}
