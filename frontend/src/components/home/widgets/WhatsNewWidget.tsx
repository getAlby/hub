import { compare } from "compare-versions";
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
          <p className="font-semibold mt-2 text-sm">
            Make sure to update! you're currently running {info.version}
          </p>
        )}
      </CardContent>
    </Card>
  );
}
