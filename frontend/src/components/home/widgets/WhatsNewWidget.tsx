import { compare } from "compare-versions";
import { ExternalLinkButton } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useInfo } from "src/hooks/useInfo";

export function WhatsNewWidget() {
  const { data: info } = useInfo();
  const { data: albyInfo } = useAlbyInfo();

  if (!info || !albyInfo || !albyInfo.hub.latestReleaseNotes) {
    return null;
  }

  const upToDate =
    info.version &&
    info.version.startsWith("v") &&
    compare(info.version.substring(1), albyInfo.hub.latestVersion, ">=");

  return (
    <Card>
      <CardHeader>
        <CardTitle>What's New in {upToDate && "your "}Alby Hub?</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-xl font-semibold">v{albyInfo.hub.latestVersion}</p>
        <p className="text-muted-foreground mt-1">
          {albyInfo.hub.latestReleaseNotes}
        </p>
        {!upToDate && (
          <div className="mt-4 flex gap-2 items-center">
            <p className="text-sm">You're currently running {info.version}</p>
            <ExternalLinkButton
              to={`https://getalby.com/update/hub?version=${info.version}`}
              size="sm"
            >
              Update Now
            </ExternalLinkButton>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
