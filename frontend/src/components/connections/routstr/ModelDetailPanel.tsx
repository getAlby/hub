import { useEffect, useState } from "react";
import { ArrowRightIcon, ExternalLinkIcon, CopyIcon } from "lucide-react";
import type { RoutstrdModel } from "src/hooks/useRoutstrd";
import {
  getModelDisplayName,
  detectAuthor,
  formatSatsPerM,
} from "./ModelSelectUtils";
import { getRoutstrdModelProviders } from "src/hooks/useRoutstrd";
import { toast } from "sonner";
import { copyToClipboard } from "src/lib/clipboard";

type Props = {
  model: RoutstrdModel;
  onSelect: () => void;
  /** Hide Select CTA when browsing catalog only */
  browseOnly?: boolean;
};

function fmtContext(cl?: number | null): string {
  if (!cl) {
    return "—";
  }
  if (cl >= 1_000_000) {
    return (cl / 1_000_000).toFixed(0) + "M";
  }
  if (cl >= 1_000) {
    return Math.round(cl / 1_000).toLocaleString() + "K";
  }
  return cl.toLocaleString();
}

export default function ModelDetailPanel({
  model,
  onSelect,
  browseOnly,
}: Props) {
  const name = getModelDisplayName(model);
  const { author } = detectAuthor(model);
  const description = model.description || "";
  const arch = model.architecture || {};
  const modelUrl = `https://routstr.com/models/${encodeURIComponent(model.id || model.name || "")}`;

  // Fetch cheapest provider pricing (like ModelPricingStrip does)
  const [cheapestPrompt, setCheapestPrompt] = useState<number | null>(null);
  const [cheapestCompletion, setCheapestCompletion] = useState<number | null>(
    null
  );
  const [providerCount, setProviderCount] = useState(0);
  const [pricingLoading, setPricingLoading] = useState(false);

  useEffect(() => {
    if (!model.id) {
      return;
    }
    let cancelled = false;
    setPricingLoading(true);
    getRoutstrdModelProviders(model.id)
      .then((result) => {
        if (cancelled) {
          return;
        }
        const provs = result?.providers || [];
        setProviderCount(provs.length);
        if (provs.length > 0) {
          // Mirror ModelPricingStrip: sort by prompt+completion so the
          // cheapest provider's pricing is shown, not the first in the list.
          const cheapest = [...provs].sort(
            (a, b) =>
              (a.pricing?.prompt ?? 0) +
              (a.pricing?.completion ?? 0) -
              ((b.pricing?.prompt ?? 0) + (b.pricing?.completion ?? 0))
          )[0];
          const p = cheapest.pricing || {};
          setCheapestPrompt(p.prompt ?? null);
          setCheapestCompletion(p.completion ?? null);
        } else {
          setCheapestPrompt(null);
          setCheapestCompletion(null);
        }
        setPricingLoading(false);
      })
      .catch(() => {
        if (!cancelled) {
          setPricingLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [model.id]);

  const handleCopyId = () => {
    if (model.id) {
      copyToClipboard(model.id);
      toast.success("Model ID copied");
    }
  };

  const hasCheapest = cheapestPrompt !== null && cheapestCompletion !== null;

  return (
    <div className="flex flex-col h-full">
      {/* Header: model name + provider */}
      <div>
        <h2 className="text-base font-semibold text-foreground leading-tight">
          {name}
        </h2>
        {author && (
          <p className="text-[11px] text-muted-foreground/60 mt-0.5">
            by {author}
          </p>
        )}
      </div>

      {/* Description */}
      {description && (
        <p className="text-xs text-muted-foreground/70 leading-relaxed mt-3 line-clamp-4">
          {description}
        </p>
      )}

      {/* Divider */}
      <div className="mt-3 mb-2 border-t border-border/40" />

      {/* Stats — cheapest provider pricing */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-[11px] text-muted-foreground/50">Context</span>
          <span className="text-sm text-foreground/90 tabular-nums font-medium">
            {fmtContext(model.context_length)}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-[11px] text-muted-foreground/50">Input</span>
          <span className="text-sm text-foreground/90 tabular-nums font-medium">
            {pricingLoading ? (
              <span className="text-xs text-muted-foreground/50">
                Loading...
              </span>
            ) : hasCheapest ? (
              `${formatSatsPerM(cheapestPrompt)} sats/1M tokens`
            ) : (
              "—"
            )}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-[11px] text-muted-foreground/50">Output</span>
          <span className="text-sm text-foreground/90 tabular-nums font-medium">
            {pricingLoading ? (
              <span className="text-xs text-muted-foreground/50">
                Loading...
              </span>
            ) : hasCheapest ? (
              `${formatSatsPerM(cheapestCompletion)} sats/1M tokens`
            ) : (
              "—"
            )}
          </span>
        </div>
        {providerCount > 1 && (
          <div className="flex items-center justify-between">
            <span className="text-[11px] text-muted-foreground/50">
              Providers
            </span>
            <span className="text-sm text-foreground/90 tabular-nums font-medium">
              {providerCount}
            </span>
          </div>
        )}
        {arch.modality && (
          <div className="flex items-center justify-between">
            <span className="text-[11px] text-muted-foreground/50">
              Modality
            </span>
            <span className="text-xs text-foreground/80 tabular-nums font-mono">
              {arch.modality}
            </span>
          </div>
        )}
        {model.created && (
          <div className="flex items-center justify-between">
            <span className="text-[11px] text-muted-foreground/50">
              Created
            </span>
            <span className="text-xs text-foreground/80 tabular-nums">
              {new Date(model.created * 1000).toLocaleDateString("en-US", {
                year: "numeric",
                month: "short",
              })}
            </span>
          </div>
        )}
      </div>

      {/* Spacer */}
      <div className="flex-1 min-h-2" />

      {/* Model ID + Copy */}
      <div className="flex items-center gap-2 mb-3">
        <code className="text-[11px] text-muted-foreground/60 font-mono truncate flex-1">
          {model.id}
        </code>
        <button
          type="button"
          onClick={handleCopyId}
          className="shrink-0 inline-flex items-center justify-center h-6 w-6 rounded hover:bg-accent transition-colors"
          title="Copy model ID"
        >
          <CopyIcon className="h-3 w-3 text-muted-foreground/50" />
        </button>
      </div>

      {/* Bottom: routstr link + Select */}
      <div className="flex items-center justify-between gap-2 pt-2 border-t border-border/40">
        <a
          href={modelUrl}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(e) => e.stopPropagation()}
          className="inline-flex items-center gap-1 text-[10px] text-muted-foreground/50 hover:text-foreground transition-colors"
        >
          <ExternalLinkIcon className="h-3 w-3" />
          View on routstr.com
        </a>
        {!browseOnly && (
          <button
            type="button"
            onClick={onSelect}
            className="inline-flex items-center justify-center rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors px-3 py-1.5 text-xs font-medium gap-1"
          >
            Select
            <ArrowRightIcon className="h-3.5 w-3.5" />
          </button>
        )}
      </div>
    </div>
  );
}
