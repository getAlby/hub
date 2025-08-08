import { AlertTriangleIcon, CheckCircle2Icon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useHealthCheck } from "src/hooks/useHealthCheck";
import { AlbyInfoIncident, HealthAlarm } from "src/types";
import { ExternalLinkButton } from "../ui/custom/external-link-button";

export function HealthCheckAlert() {
  const { data: health } = useHealthCheck();

  const ok = !health?.alarms?.length;

  if (!health) {
    return null;
  }

  if (ok) {
    return null;
  }

  function getAlarmTitle(alarm: HealthAlarm) {
    // TODO: could show extra data from alarm.rawDetails
    // for some alarm types
    try {
      switch (alarm.kind) {
        case "alby_service":
          return (
            "Alby Services: " +
            (alarm.rawDetails as AlbyInfoIncident[])
              ?.map((incident) => `${incident.name} (${incident.status})`)
              .join(", ")
          );
        case "channels_offline":
          return "One or more channels are offline";
        case "node_not_ready":
          return "Node is not ready";
        case "nostr_relay_offline":
          return "Could not connect to relay";
        case "vss_no_subscription":
          return "Your lightning channel data is stored encrypted by Alby's Versioned Storage Service which is a paid feature. Restart your subscription or send your funds to another wallet as soon as possible.";
      }
    } catch (error) {
      console.error("failed to parse alarm details", alarm.kind, error);
    }
    return alarm.kind || "Unknown";
  }

  return (
    <>
      <Alert className="animate-highlight">
        {ok ? (
          <>
            <CheckCircle2Icon className="h-4 w-4" />
            <AlertTitle>Alby Hub is running smoothly</AlertTitle>
          </>
        ) : (
          <>
            <AlertTriangleIcon className="h-4 w-4" />
            <AlertTitle>
              {health.alarms.length} issues impacting your hub were found
            </AlertTitle>
          </>
        )}
        <AlertDescription>
          {health.alarms?.length && (
            <ul className="mt-2 whitespace-pre-wrap list-disc list-inside">
              {health.alarms?.map((alarm) => (
                <li key={alarm.kind}>{getAlarmTitle(alarm)}</li>
              ))}
            </ul>
          )}
          <div className="mt-4 flex gap-2">
            <ExternalLinkButton
              size="sm"
              variant="secondary"
              to="https://status.getalby.com"
            >
              Alby Service Status
            </ExternalLinkButton>
            <ExternalLinkButton
              size="sm"
              variant="secondary"
              to="https://guides.getalby.com/user-guide/alby-hub/faq/what-happens-if-the-healthcheck-indicator-turns-red"
            >
              Learn More
            </ExternalLinkButton>
          </div>
        </AlertDescription>
      </Alert>
    </>
  );
}
