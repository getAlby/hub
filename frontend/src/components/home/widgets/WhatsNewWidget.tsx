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

  if (
    !info ||
    !albyInfo ||
    !albyInfo.hub.latestReleaseNotes ||
    info.hideUpdateBanner
  ) {
    return null;
  }

  // Same normalization as useBanner.tsx: strip leading "v"/whitespace on
  // both sides and never throw, so an unparseable version cannot render a
  // false-positive "Update Now" button (#1870).
  let upToDate: boolean;
  try {
    const current = info.version?.trim().replace(/^v/i, "");
    const latest = albyInfo.hub.latestVersion?.trim().replace(/^v/i, "");
    upToDate =
      Boolean(current) && Boolean(latest) && compare(current, latest, ">=");
  } catch {
    upToDate = true;
  }

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
