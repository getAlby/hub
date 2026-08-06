import { TriangleAlertIcon } from "lucide-react";
import { Link } from "react-router";
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { getConnectionIssueCopy } from "src/lib/connectionIssues";
import { ConnectionIssue } from "src/types";

export function ConnectionIssueAlert({
  appName,
  issue,
  onViewDetails,
  showTimestamp = true,
}: {
  appName: string;
  issue: ConnectionIssue;
  onViewDetails: () => void;
  showTimestamp?: boolean;
}) {
  const copy = getConnectionIssueCopy(appName, issue, onViewDetails);

  return (
    <Alert variant="warning">
      <TriangleAlertIcon />
      <AlertTitle className="line-clamp-none">{copy.title}</AlertTitle>
      <AlertDescription>
        <p>{copy.description}</p>
        <p className="font-mono text-xs break-all">
          {issue.errorCode}: {issue.errorMessage}
        </p>
        {showTimestamp && (
          <p className="text-xs">
            {new Date(issue.createdAt).toLocaleString()}
          </p>
        )}
      </AlertDescription>
      <AlertAction>
        {copy.href ? (
          <Button asChild size="sm" variant="secondary">
            <Link to={copy.href}>{copy.action}</Link>
          </Button>
        ) : (
          <Button size="sm" variant="secondary" onClick={copy.onClick}>
            {copy.action}
          </Button>
        )}
      </AlertAction>
    </Alert>
  );
}

export function ConnectionIssuesCard({
  appName,
  issues,
  onViewDetails,
}: {
  appName: string;
  issues: ConnectionIssue[] | undefined;
  onViewDetails: () => void;
}) {
  if (!issues?.length) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Connection Issues</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3">
        {issues.map((issue) => (
          <ConnectionIssueAlert
            key={issue.id}
            appName={appName}
            issue={issue}
            onViewDetails={onViewDetails}
          />
        ))}
      </CardContent>
    </Card>
  );
}
