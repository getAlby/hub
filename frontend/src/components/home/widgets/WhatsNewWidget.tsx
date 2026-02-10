import { compare } from "compare-versions";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
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
        <CardDescription>{albyInfo.hub.latestReleaseNotes}</CardDescription>
      </CardHeader>
      {!upToDate && (
        <CardContent className="text-right">
          <ExternalLinkButton
            to={`https://getalby.com/update/hub?version=${info.version}`}
            size="sm"
          >
            Update Now
          </ExternalLinkButton>
        </CardContent>
      )}
    </Card>
  );
}
