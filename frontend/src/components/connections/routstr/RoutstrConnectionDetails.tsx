import { useState } from "react";
import { CopyIcon, EyeIcon, EyeOffIcon } from "lucide-react";
import { toast } from "sonner";
import { Button } from "src/components/ui/button";
import { Label } from "src/components/ui/label";
import { copyToClipboard } from "src/lib/clipboard";
import { cn } from "src/lib/utils";
import {
  ROUTSTR_LOCAL_ENDPOINT,
  getRoutstrHubEndpoint,
  isHubOpenedLocally,
} from "./constants";

type Props = {
  apiKey: string;
  /** When false, Hub URL row is hidden (pre-proxy). Default true. */
  showHubUrl?: boolean;
  className?: string;
};

function CopyRow({
  label,
  value,
  displayValue,
  helper,
  emphasize,
  onExtraAction,
}: {
  label: string;
  value: string;
  displayValue?: string;
  helper?: string;
  emphasize?: boolean;
  onExtraAction?: { icon: React.ReactNode; onClick: () => void; title: string };
}) {
  const handleCopy = () => {
    copyToClipboard(value);
    toast.success("Copied");
  };

  return (
    <div
      className={cn(
        "rounded-lg border p-3 space-y-1.5",
        emphasize
          ? "border-primary/40 bg-primary/5"
          : "border-border bg-muted/40"
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1 space-y-1">
          <Label className="text-xs text-muted-foreground font-medium">
            {label}
            {emphasize && (
              <span className="ml-2 text-[10px] uppercase tracking-wide text-primary">
                Recommended
              </span>
            )}
          </Label>
          <div className="font-mono text-sm break-all text-foreground">
            {displayValue ?? value}
          </div>
          {helper && (
            <p className="text-xs text-muted-foreground leading-snug">
              {helper}
            </p>
          )}
        </div>
        <div className="flex items-center gap-1 shrink-0">
          {onExtraAction && (
            <button
              type="button"
              onClick={onExtraAction.onClick}
              className="inline-flex items-center justify-center h-8 w-8 rounded-lg hover:bg-accent transition-colors"
              title={onExtraAction.title}
            >
              {onExtraAction.icon}
            </button>
          )}
          <Button
            type="button"
            size="sm"
            variant="outline"
            className="shrink-0"
            onClick={handleCopy}
          >
            <CopyIcon className="h-3.5 w-3.5" />
            Copy
          </Button>
        </div>
      </div>
    </div>
  );
}

export function RoutstrConnectionDetails({
  apiKey,
  showHubUrl = true,
  className,
}: Props) {
  const hubUrl = getRoutstrHubEndpoint();
  const preferLocal = isHubOpenedLocally();
  const [apiKeyVisible, setApiKeyVisible] = useState(false);

  const blindedKey =
    apiKey.length > 12
      ? `${apiKey.slice(0, 6)}...${apiKey.slice(-4)}`
      : `${apiKey.slice(0, 4)}...`;

  return (
    <div className={cn("space-y-3", className)}>
      {/* API Key with blind/unblind */}
      <CopyRow
        label="API Key"
        value={apiKey}
        displayValue={apiKeyVisible ? apiKey : blindedKey}
        onExtraAction={{
          icon: apiKeyVisible ? (
            <EyeOffIcon className="h-4 w-4 text-muted-foreground/70" />
          ) : (
            <EyeIcon className="h-4 w-4 text-muted-foreground/70" />
          ),
          onClick: () => setApiKeyVisible((v) => !v),
          title: apiKeyVisible ? "Hide API key" : "Show API key",
        }}
      />

      {/* Base URL — single section with two sub-entries */}
      <div
        className={cn(
          "rounded-lg border p-3 space-y-3",
          preferLocal
            ? "border-primary/40 bg-primary/5"
            : "border-border bg-muted/40"
        )}
      >
        <Label className="text-xs text-muted-foreground font-medium">
          Base URL
          {preferLocal && (
            <span className="ml-2 text-[10px] uppercase tracking-wide text-primary">
              Recommended
            </span>
          )}
        </Label>

        {/* Local */}
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground font-medium">
              Same device
            </p>
            <p className="font-mono text-sm break-all text-foreground">
              {ROUTSTR_LOCAL_ENDPOINT}
            </p>
          </div>
          <Button
            type="button"
            size="sm"
            variant="outline"
            className="shrink-0"
            onClick={() => {
              copyToClipboard(ROUTSTR_LOCAL_ENDPOINT);
              toast.success("Copied");
            }}
          >
            <CopyIcon className="h-3.5 w-3.5" />
            Copy
          </Button>
        </div>

        {/* External */}
        {showHubUrl && (
          <div className="flex items-start justify-between gap-2 pt-2 border-t border-border/40">
            <div className="min-w-0 flex-1">
              <p className="text-xs text-muted-foreground font-medium">
                External device
              </p>
              <p className="font-mono text-sm break-all text-foreground">
                {hubUrl}
              </p>
            </div>
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="shrink-0"
              onClick={() => {
                copyToClipboard(hubUrl);
                toast.success("Copied");
              }}
            >
              <CopyIcon className="h-3.5 w-3.5" />
              Copy
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
