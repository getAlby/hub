import { useEffect, useState } from "react";
import { ChevronDown, ChevronUp, ZapIcon, CheckIcon } from "lucide-react";
import { cn } from "src/lib/utils";
import { getRoutstrdModelProviders } from "src/hooks/useRoutstrd";
import { formatSatsPerM } from "./ModelSelectUtils";

type ProviderInfo = {
  baseUrl: string;
  pricing: {
    prompt: number;
    completion: number;
    request: number;
    max_cost: number;
  };
};

type Props = {
  modelId: string;
  modelName?: string;
  selectedProvider?: string;
  onProviderChange?: (baseUrl: string | null) => void;
  readOnly?: boolean;
};

function prettyProviderName(url: string): string {
  const cleaned = url
    .replace(/^https?:\/\//, "")
    .replace(/\/$/, "")
    .toLowerCase();
  // Known provider patterns
  const known: Record<string, string> = {
    "node1.routstr.blazelight.dev": "Blazelight",
    "llm.satsandsports.cash": "SatsAndSports",
    "routstr.otrta.me": "Otrta",
    "routstr.satoshisend.xyz": "SatoshiSend",
    "api.nonkycai.com": "NonKYC AI",
    "staging.routstr.com": "Routstr",
    "ai.redsh1ft.com": "Redsh1ft",
    "privateprovider.xyz": "Private",
    "routstr.cypherpunk.xyz": "Cypherpunk",
    "api.plebchat.com": "PlebChat",
    "routstr.githappens.xyz": "GitHappens",
    "rossi.routstr.com": "Rossi",
  };
  return (
    known[cleaned] ||
    cleaned.split(".").slice(-2, -1)[0] ||
    cleaned.slice(0, 15)
  );
}

export default function ModelPricingStrip({
  modelId,
  selectedProvider,
  onProviderChange,
  readOnly,
}: Props) {
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [localSelected, setLocalSelected] = useState<string | null>(null);

  const activeProvider = selectedProvider ?? localSelected;

  useEffect(() => {
    if (!modelId) {
      setProviders([]);
      setExpanded(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    getRoutstrdModelProviders(modelId)
      .then((result) => {
        if (cancelled) {
          return;
        }
        const provs = result?.providers || [];
        // Re-sort by prompt+completion sum per million tokens — matches routstrd's internal ranking
        const sorted = [...provs].sort((a, b) => {
          const aTotal =
            (a.pricing?.prompt ?? 0) + (a.pricing?.completion ?? 0);
          const bTotal =
            (b.pricing?.prompt ?? 0) + (b.pricing?.completion ?? 0);
          return aTotal - bTotal;
        });
        setProviders(sorted);
        setLoading(false);
      })
      .catch(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [modelId]);

  useEffect(() => {
    setExpanded(false);
    setLocalSelected(null);
  }, [modelId]);

  const handleSelect = (baseUrl: string) => {
    if (readOnly) {
      return;
    }
    const next = activeProvider === baseUrl ? null : baseUrl;
    setLocalSelected(next);
    onProviderChange?.(next);
  };

  if (loading) {
    return (
      <div className="flex items-center gap-1.5 px-1 py-1">
        <span className="h-3 w-3 rounded-full border border-border animate-spin border-t-transparent" />
        <span className="text-[10px] text-muted-foreground/60">
          Loading pricing...
        </span>
      </div>
    );
  }

  if (providers.length === 0) {
    return null;
  }

  const cheapest = providers[0];
  const promptPrice = cheapest?.pricing?.prompt;
  const completionPrice = cheapest?.pricing?.completion;
  const isFree = promptPrice === 0 && completionPrice === 0;

  return (
    <div className="mt-1">
      {/* Compact pricing summary */}
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className={cn(
          "flex items-center gap-1.5 px-3 py-2 rounded-lg w-full text-left border border-border/50",
          "hover:bg-accent/30 transition-colors",
          expanded && "bg-accent/20 border-primary/20"
        )}
      >
        <ZapIcon className="h-4 w-4 text-amber-500 shrink-0" />
        {isFree ? (
          <span className="text-sm text-muted-foreground/80">Free model</span>
        ) : (
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <span className="text-xs text-muted-foreground/80">
              from{" "}
              <strong className="text-foreground/90 font-semibold text-sm">
                {formatSatsPerM(promptPrice)} in /{" "}
                {formatSatsPerM(completionPrice)} out
              </strong>
            </span>
            <span className="text-[10px] text-muted-foreground/40 mx-0.5">
              ·
            </span>
            <span className="text-[10px] text-muted-foreground/50">
              {providers.length} providers
            </span>
          </div>
        )}
        {activeProvider && !readOnly && (
          <span className="text-[10px] text-primary font-medium shrink-0">
            {prettyProviderName(activeProvider)}
          </span>
        )}
        {activeProvider && readOnly && (
          <span className="text-[10px] text-primary font-medium shrink-0">
            Pinned: {prettyProviderName(activeProvider)}
          </span>
        )}
        {!activeProvider && readOnly && (
          <span className="text-[10px] text-muted-foreground/60 font-medium shrink-0">
            Auto-routing
          </span>
        )}
        {providers.length > 1 && (
          <span className="ml-auto shrink-0">
            {expanded ? (
              <ChevronUp className="h-3.5 w-3.5 text-muted-foreground/40" />
            ) : (
              <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/40" />
            )}
          </span>
        )}
      </button>

      {/* Expanded provider list */}
      {expanded && (
        <div className="mt-1.5 rounded-lg border border-border/50 bg-muted/10 overflow-hidden">
          <div className="px-3 py-2.5 border-b border-border/30 bg-muted/20">
            <p className="text-xs text-muted-foreground/80 leading-relaxed">
              {readOnly
                ? activeProvider
                  ? `Provider: ${prettyProviderName(activeProvider)}`
                  : "Auto-routing: Cheapest provider selected at request time."
                : activeProvider
                  ? `Requests will start with ${prettyProviderName(activeProvider)}. Automatically failover if unavailable.`
                  : "Routstr will use the cheapest available provider at request time."}
              {!readOnly &&
                " Click a provider to select, or deselect for auto-routing."}
            </p>
          </div>
          <div className="divide-y divide-border/20">
            {providers.map((p, i) => {
              const isCheapest = i === 0;
              const isSelected = activeProvider === p.baseUrl;
              const pName = prettyProviderName(p.baseUrl);
              const RowTag = readOnly ? "div" : "button";
              return (
                <RowTag
                  key={p.baseUrl}
                  {...(readOnly
                    ? {}
                    : {
                        type: "button" as const,
                        onClick: () => handleSelect(p.baseUrl),
                      })}
                  className={cn(
                    "flex items-center gap-3 w-full px-3 py-2.5 text-left transition-colors",
                    !readOnly && "hover:bg-accent/30",
                    isSelected && "bg-primary/5"
                  )}
                >
                  {/* Radio indicator — hidden in readOnly */}
                  {!readOnly && (
                    <div
                      className={cn(
                        "shrink-0 flex items-center justify-center h-5 w-5 rounded-full border-2 transition-colors",
                        isSelected
                          ? "border-primary bg-primary text-primary-foreground"
                          : "border-muted-foreground/30"
                      )}
                    >
                      {isSelected && <CheckIcon className="h-3 w-3" />}
                    </div>
                  )}

                  {/* Provider info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1.5">
                      <span
                        className={cn(
                          "text-xs font-medium",
                          isSelected ? "text-foreground" : "text-foreground/80"
                        )}
                      >
                        {pName}
                      </span>
                      {isCheapest && !isSelected && (
                        <span className="inline-flex items-center px-1.5 py-0.5 rounded-full bg-amber-500/10 text-[9px] font-medium text-amber-600 dark:text-amber-400">
                          Best price
                        </span>
                      )}
                      {isCheapest && isSelected && !readOnly && (
                        <span className="inline-flex items-center px-1.5 py-0.5 rounded-full bg-primary/10 text-[9px] font-medium text-primary">
                          Recommended
                        </span>
                      )}
                      {isSelected && readOnly && (
                        <span className="inline-flex items-center px-1.5 py-0.5 rounded-full bg-primary/10 text-[9px] font-medium text-primary">
                          Pinned
                        </span>
                      )}
                    </div>
                    <div className="text-[10px] text-muted-foreground/50 font-mono truncate mt-0.5">
                      {p.baseUrl.replace(/^https?:\/\//, "").replace(/\/$/, "")}
                    </div>
                  </div>

                  {/* Pricing */}
                  <div className="shrink-0 text-right">
                    <div className="text-xs tabular-nums font-medium text-foreground/90">
                      {formatSatsPerM(p.pricing?.prompt)} /{" "}
                      {formatSatsPerM(p.pricing?.completion)}
                    </div>
                    <div className="text-[9px] text-muted-foreground/40">
                      sats/1M tokens
                    </div>
                  </div>
                </RowTag>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
