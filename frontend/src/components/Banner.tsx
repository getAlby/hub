import { XIcon } from "lucide-react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";

import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export function Banner({ onDismiss }: { onDismiss: () => void }) {
  const { data: info } = useInfo();
  const { data: albyInfo } = useAlbyInfo();
  const { data: albyMe } = useAlbyMe();

  if (!info || !albyInfo) {
    return null;
  }

  // vss migration (alby cloud only)
  // TODO: remove after 2026-01-01
  const vssMigrationRequired =
    info.oauthRedirect &&
    !!albyMe?.subscription.plan_code.includes("buzz") &&
    info.vssSupported &&
    !info.ldkVssEnabled;

  if (vssMigrationRequired) {
    return (
      <div className="fixed w-full bg-foreground text-background z-20 py-2 text-sm flex items-center justify-center">
        <Link
          to={`/settings/backup?dynamic=true`}
          className="w-full px-12 md:px-24"
        >
          <p className="line-clamp-2 md:block whitespace-normal md:whitespace-nowrap overflow-hidden text-ellipsis text-center">
            <span className="font-semibold mr-2">
              Enable Dynamic Channels Backup
            </span>
            <span>•</span>
            <span className="ml-2">Upgrade your channel backups now</span>
          </p>
        </Link>
        <XIcon
          className="absolute right-4 cursor-pointer w-4 text-background"
          onClick={onDismiss}
        />
      </div>
    );
  }

  return (
    <div className="fixed w-full bg-foreground text-background z-20 py-2 text-sm flex items-center justify-center">
      <ExternalLink
        to={`https://getalby.com/update/hub?version=${info?.version}`}
        className="w-full px-12 md:px-24"
      >
        <p className="line-clamp-2 md:block whitespace-normal md:whitespace-nowrap overflow-hidden text-ellipsis text-center">
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
