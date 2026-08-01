import { CopyIcon, ExternalLinkIcon, KeyIcon, ShieldIcon } from "lucide-react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import SettingsHeader from "src/components/SettingsHeader";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useApps } from "src/hooks/useApps";
import { copyToClipboard } from "src/lib/clipboard";
import type { App } from "src/types";

const ROUTSTR_APP_ID = "routstr";

type RoutstrMetadata = {
  apiKey?: string;
  modelId?: string;
  clientId?: string;
  balance?: number;
};

export default function RoutstrApiKeys() {
  const navigate = useNavigate();
  const { data: appsData, isLoading } = useApps(undefined, undefined, {
    appStoreAppId: ROUTSTR_APP_ID,
  });

  const routstrApps = (appsData?.apps || []).filter(
    (
      app
    ): app is App & {
      metadata: { [key: string]: unknown; routstr: RoutstrMetadata };
    } => !!(app.metadata as Record<string, unknown>)?.routstr
  );

  const totalKeys = routstrApps.length;
  const totalBalance = routstrApps.reduce(
    (sum, app) => sum + (app.metadata.routstr.balance || 0),
    0
  );

  const handleCopyKey = (key: string) => {
    copyToClipboard(key);
    toast.success("API key copied to clipboard");
  };

  const handleViewConnection = (appId: number) => {
    navigate(`/apps/${appId}`);
  };

  return (
    <>
      <SettingsHeader
        pageTitle="Routstr API Keys"
        title="Routstr API Keys"
        description={
          <>
            Universal API keys for Routstr across all connections. Each key
            works with all models — choose the model in your AI client. Keys are
            included in your{" "}
            <button
              type="button"
              onClick={() => navigate("/settings/backup")}
              className="text-foreground underline underline-offset-2 hover:no-underline"
            >
              Hub backup
            </button>
            . Endpoints:{" "}
            <code className="text-xs">http://localhost:8008/v1</code> (same
            device) or <code className="text-xs">/routstr/v1</code> on this Hub
            (remote).
          </>
        }
      />

      {/* Summary card */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <KeyIcon className="h-5 w-5" />
              <CardTitle className="text-lg">Overview</CardTitle>
            </div>
            <Badge variant={totalKeys > 0 ? "positive" : "secondary"}>
              {totalKeys} key{totalKeys !== 1 ? "s" : ""}
            </Badge>
          </div>
          <CardDescription>
            {totalKeys > 0
              ? `${totalKeys} Routstr API key${totalKeys !== 1 ? "s are" : " is"} active across your connections with ${totalBalance} total sats balance.`
              : "No Routstr API keys found. Create one from a Routstr connection."}
          </CardDescription>
        </CardHeader>
      </Card>

      {/* Key list */}
      {isLoading ? (
        <div className="text-sm text-muted-foreground py-4">
          Loading keys...
        </div>
      ) : routstrApps.length === 0 ? (
        <Card>
          <CardContent className="py-8 text-center text-sm text-muted-foreground">
            <KeyIcon className="h-8 w-8 mx-auto mb-2 opacity-40" />
            <p>No API keys yet</p>
            <p className="text-xs mt-1">
              Create an API key from the Routstr connection to see it here.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {routstrApps.map((app) => {
            const meta = app.metadata.routstr;
            return (
              <Card key={app.id}>
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <KeyIcon className="h-4 w-4 text-muted-foreground" />
                      <CardTitle className="text-sm font-medium">
                        {meta.modelId || "Unknown Model"}
                      </CardTitle>
                    </div>
                    <Badge
                      variant={
                        meta.balance && meta.balance > 0
                          ? "positive"
                          : "secondary"
                      }
                    >
                      {meta.balance ?? "?"} sats
                    </Badge>
                  </div>
                  <CardDescription className="text-xs">
                    Connection: {app.name}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  {/* Key display */}
                  {meta.apiKey && (
                    <div className="flex items-center gap-2">
                      <code className="flex-1 text-xs bg-muted p-2 rounded truncate font-mono">
                        {meta.apiKey.slice(0, 24)}...
                      </code>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleCopyKey(meta.apiKey!)}
                        title="Copy API key"
                      >
                        <CopyIcon className="h-3 w-3" />
                      </Button>
                    </div>
                  )}

                  {/* Actions */}
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => handleViewConnection(app.id)}
                    >
                      <ExternalLinkIcon className="h-3 w-3 mr-1" />
                      View Connection
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Backup info card */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <ShieldIcon className="h-5 w-5" />
            <CardTitle className="text-lg">Backup</CardTitle>
          </div>
          <CardDescription>
            Your Routstr API keys are stored as part of each connection's
            metadata in the Hub database. Performing a full Hub backup includes
            all your API keys.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant="outline"
            onClick={() => navigate("/settings/backup")}
          >
            <ShieldIcon className="h-3 w-3 mr-1" />
            Go to Backup
          </Button>
        </CardContent>
      </Card>
    </>
  );
}
