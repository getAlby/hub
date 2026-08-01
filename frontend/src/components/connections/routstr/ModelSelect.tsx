import {
  useState,
  useRef,
  useEffect,
  useMemo,
  useCallback,
  type KeyboardEvent,
} from "react";
import {
  Search,
  X,
  PinIcon,
  PinOffIcon,
  ChevronDown,
  FilterIcon,
  ExternalLinkIcon,
} from "lucide-react";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from "src/components/ui/dialog";
import { cn } from "src/lib/utils";
import type { RoutstrdModel } from "src/hooks/useRoutstrd";
import {
  getModelDisplayName,
  isFreeModel,
  hasModality,
  togglePin,
  getPinnedModels,
  groupModelsByMonth,
} from "./ModelSelectUtils";
import ModelPricingStrip from "./ModelPricingStrip";
import ModelDetailPanel from "./ModelDetailPanel";

// ─── Props ──────────────────────────────────────────────

type Props = {
  models: RoutstrdModel[];
  value?: string;
  onChange?: (modelId: string) => void;
  disabled?: boolean;
  inline?: boolean;
  /** Browse catalog only — no selection required (key is model-universal). */
  browseOnly?: boolean;
  /** Show model count text (default true). */
  showCount?: boolean;
  /** Fires when the dialog opens — parent can refresh data silently. */
  onOpen?: () => void;
};

// ─── Selected Model Chip ─────────────────────────────────

function SelectedChip({
  model,
  onRemove,
}: {
  model: RoutstrdModel;
  onRemove: () => void;
}) {
  const displayName = getModelDisplayName(model);

  return (
    <span className="inline-flex items-center gap-1.5 h-8 pl-2 pr-1 rounded-full border border-border bg-muted/50 text-sm cursor-default select-none">
      <span className="max-w-[180px] truncate font-medium">{displayName}</span>
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          onRemove();
        }}
        className="ml-0.5 inline-flex items-center justify-center h-5 w-5 rounded-full hover:bg-muted-foreground/20 transition-colors"
        aria-label="Remove model"
      >
        <X className="h-3 w-3" />
      </button>
    </span>
  );
}

// ─── Modality Popup Filter (multi-select checkboxes) ──

type PopupState = { anchor: "input" | "output" | "month" } | null;

const MODALITY_OPTIONS = ["text", "image", "file", "audio", "video"];

function ModalityPopup({
  selected,
  onToggle,
  onClose,
}: {
  selected: string[];
  onToggle: (value: string) => void;
  onClose: () => void;
}) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [onClose]);

  return (
    <div
      ref={ref}
      className="absolute top-full left-0 mt-1 z-50 p-2 rounded-lg border border-border bg-popover shadow-md min-w-[140px]"
    >
      {MODALITY_OPTIONS.map((mod) => (
        <label
          key={mod}
          className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-accent cursor-pointer text-xs"
        >
          <input
            type="checkbox"
            checked={selected.includes(mod)}
            onChange={() => onToggle(mod)}
            className="rounded border-border"
          />
          <span className="capitalize">{mod}</span>
        </label>
      ))}
    </div>
  );
}

// ─── Month Filter Popup ──────────────────────────────────

function MonthPopup({
  months,
  selected,
  onToggle,
  onClose,
}: {
  months: string[];
  selected: string[];
  onToggle: (value: string) => void;
  onClose: () => void;
}) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [onClose]);

  return (
    <div
      ref={ref}
      className="absolute top-full left-0 mt-1 z-50 p-2 rounded-lg border border-border bg-popover shadow-md min-w-[150px] max-h-[200px] overflow-y-auto"
    >
      {months.map((month) => (
        <label
          key={month}
          className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-accent cursor-pointer text-xs"
        >
          <input
            type="checkbox"
            checked={selected.includes(month)}
            onChange={() => onToggle(month)}
            className="rounded border-border"
          />
          <span>{month}</span>
        </label>
      ))}
    </div>
  );
}

// ─── Modality Filter Button ──────────────────────────────

function ModalityFilterButton({
  label,
  selected,
  onClick,
}: {
  label: string;
  selected: string[];
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs transition-colors",
        selected.length > 0
          ? "bg-primary/10 text-primary"
          : "bg-muted/60 text-muted-foreground hover:text-foreground"
      )}
    >
      {label}
      {selected.length > 0 && (
        <span className="inline-flex items-center justify-center h-4 min-w-[14px] px-[3px] rounded-full bg-primary text-[9px] font-medium text-primary-foreground tabular-nums">
          {selected.length}
        </span>
      )}
      <ChevronDown className="h-3 w-3" />
    </button>
  );
}

// ─── Month Group Header ──────────────────────────────────

function MonthGroup({ label, count }: { label: string; count: number }) {
  return (
    <div className="flex items-center gap-2 px-4 py-2">
      <span className="text-[11px] font-medium text-muted-foreground/60 uppercase tracking-wider">
        {label}
      </span>
      <span className="text-[10px] text-muted-foreground/30 tabular-nums">
        {count}
      </span>
    </div>
  );
}

// ─── Model Row ──────────────────────────────────────────

function ModelRow({
  model,
  isSelected,
  pinned,
  onSelect,
  onPinToggle,
  highlighted,
  onHover,
}: {
  model: RoutstrdModel;
  isSelected: boolean;
  pinned: boolean;
  onSelect: () => void;
  onPinToggle: () => void;
  highlighted: boolean;
  onHover?: () => void;
}) {
  const displayName = getModelDisplayName(model);
  const modelUrl = `https://routstr.com/models/${encodeURIComponent(model.id || model.name || "")}`;

  return (
    <div
      role="option"
      aria-selected={highlighted}
      onClick={onSelect}
      onMouseEnter={onHover}
      className={cn(
        "flex items-center gap-2.5 px-3 py-1.5 cursor-pointer transition-colors text-sm group",
        "hover:bg-accent/60",
        highlighted && "bg-accent/40",
        isSelected && "opacity-50 pointer-events-none"
      )}
    >
      {/* Model name — clean, no badges or pricing in rows (shown in right panel) */}
      <div className="flex-1 min-w-0">
        <span className="truncate font-medium block leading-tight">
          {displayName}
        </span>
      </div>

      {/* Routstr model link */}
      <a
        href={modelUrl}
        target="_blank"
        rel="noopener noreferrer"
        onClick={(e) => e.stopPropagation()}
        className="shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
        title="View on routstr.com"
      >
        <ExternalLinkIcon className="h-3.5 w-3.5 text-muted-foreground/50 hover:text-foreground" />
      </a>

      {/* Pin button */}
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          onPinToggle();
        }}
        className={cn(
          "shrink-0 transition-all",
          pinned
            ? "opacity-100 text-primary"
            : "opacity-0 group-hover:opacity-100 text-muted-foreground/40"
        )}
        title={pinned ? "Unpin" : "Pin"}
      >
        {pinned ? (
          <PinOffIcon className="h-3.5 w-3.5" />
        ) : (
          <PinIcon className="h-3.5 w-3.5" />
        )}
      </button>
    </div>
  );
}

// ─── Main Component ─────────────────────────────────────

export default function ModelSelect({
  models,
  value,
  onChange,
  disabled,
  inline,
  browseOnly,
  showCount = true,
  onOpen,
}: Props) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [inputFilter, setInputFilter] = useState<string[]>([]);
  const [outputFilter, setOutputFilter] = useState<string[]>([]);
  const [freeOnly, setFreeOnly] = useState(false);
  const [hideUnavailable, setHideUnavailable] = useState(true);
  const [monthFilter, setMonthFilter] = useState<string[]>([]);
  const [highlightedModel, setHighlightedModel] =
    useState<RoutstrdModel | null>(null);
  const [highlightedIdx, setHighlightedIdx] = useState(0);
  const [pinnedSet, setPinnedSet] = useState<Set<string>>(
    () => new Set(getPinnedModels())
  );
  const [popup, setPopup] = useState<PopupState>(null);
  const [showFilters, setShowFilters] = useState(false);
  const searchRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const selected = useMemo(
    () => models.find((m) => m.id === value),
    [models, value]
  );

  const allMonths = useMemo(() => {
    const set = new Set<string>();
    for (const m of models) {
      if (m.created) {
        const d = new Date(m.created * 1000);
        set.add(
          `${d.toLocaleString("en-US", { month: "long" })} ${d.getFullYear()}`
        );
      }
    }
    return Array.from(set).sort(
      (a, b) => new Date(b).getTime() - new Date(a).getTime()
    );
  }, [models]);

  // Debounce search
  useEffect(() => {
    const debounceRef = window.setTimeout(
      () => setDebouncedSearch(search),
      150
    );
    return () => window.clearTimeout(debounceRef);
  }, [search]);

  // Focus search input when dialog opens
  useEffect(() => {
    if (open || inline) {
      searchRef.current?.focus();
    }
    if (!open) {
      setSearch("");
      setShowFilters(false);
    }
  }, [open, inline]);

  // Filter models
  const filtered = useMemo(() => {
    let result = models;
    if (debouncedSearch.trim()) {
      const q = debouncedSearch.toLowerCase();
      result = result.filter(
        (m) =>
          (m.id && m.id.toLowerCase().includes(q)) ||
          (m.name && m.name.toLowerCase().includes(q)) ||
          false
      );
    }
    if (inputFilter.length > 0) {
      result = result.filter((m) =>
        inputFilter.some((mod) => hasModality(m, mod, "input"))
      );
    }
    if (outputFilter.length > 0) {
      result = result.filter((m) =>
        outputFilter.some((mod) => hasModality(m, mod, "output"))
      );
    }
    if (freeOnly) {
      result = result.filter((m) => isFreeModel(m));
    }
    if (hideUnavailable) {
      // Routable = has >=1 provider with pricing. Models without sats_pricing
      // cannot be routed by the daemon (getProviderPriceRankingForModel = empty).
      result = result.filter((m) => m.sats_pricing);
    }
    if (monthFilter.length > 0) {
      result = result.filter((m) => {
        if (!m.created) {
          return false;
        }
        const d = new Date(m.created * 1000);
        const label = `${d.toLocaleString("en-US", { month: "long" })} ${d.getFullYear()}`;
        return monthFilter.includes(label);
      });
    }
    return result;
  }, [
    models,
    debouncedSearch,
    inputFilter,
    outputFilter,
    freeOnly,
    hideUnavailable,
    monthFilter,
  ]);

  // Count of routable (priced) models for the trigger label
  const routableCount = useMemo(
    () => models.filter((m) => m.sats_pricing).length,
    [models]
  );

  // Group by pinned + months
  const groupedModels = useMemo(() => {
    const pinned: RoutstrdModel[] = [];
    const rest: RoutstrdModel[] = [];
    for (const m of filtered) {
      if (pinnedSet.has(m.id)) {
        pinned.push(m);
      } else {
        rest.push(m);
      }
    }
    const sorted = [...pinned, ...rest];
    return groupModelsByMonth(sorted);
  }, [filtered, pinnedSet]);

  // Show all models (no windowing)
  const visibleSections = useMemo(() => groupedModels, [groupedModels]);

  // Flat list for keyboard nav
  const flatModels = useMemo(() => {
    const flat: RoutstrdModel[] = [];
    for (const s of visibleSections) {
      flat.push(...s.models);
    }
    return flat;
  }, [visibleSections]);

  // ── Handlers ──────────────────────────────────────────

  const handleSelect = useCallback(
    (id: string) => {
      if (browseOnly) {
        // Preview only — do not bind a model to the API key
        const m = models.find((x) => x.id === id) || null;
        setHighlightedModel(m);
        return;
      }
      onChange?.(id);
      setOpen(false);
      setHighlightedModel(null);
    },
    [onChange, browseOnly, models]
  );

  // Keyboard navigation
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setHighlightedIdx((prev) => Math.min(prev + 1, flatModels.length - 1));
        setHighlightedModel(
          flatModels[Math.min(highlightedIdx + 1, flatModels.length - 1)]
        );
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setHighlightedIdx((prev) => Math.max(prev - 1, 0));
        setHighlightedModel(flatModels[Math.max(highlightedIdx - 1, 0)]);
      } else if (e.key === "Enter" && flatModels[highlightedIdx]) {
        e.preventDefault();
        handleSelect(flatModels[highlightedIdx].id);
      }
    },
    [flatModels, highlightedIdx, handleSelect]
  );

  const handleRemove = useCallback(() => {
    onChange?.("");
    setHighlightedModel(null);
  }, [onChange]);

  const handlePinToggle = useCallback(
    (id: string) => {
      const newSet = new Set(pinnedSet);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      setPinnedSet(newSet);
      togglePin(id);
    },
    [pinnedSet]
  );

  // ── Render content (shared between dialog and inline modes) ──

  const renderContent = (closeButton?: React.ReactNode) => (
    <>
      {/* Search bar */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border">
        <Search className="h-4 w-4 text-muted-foreground shrink-0" />
        <input
          ref={searchRef}
          placeholder="Search models"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 text-sm bg-transparent outline-none placeholder:text-muted-foreground"
        />
        <span className="text-xs text-muted-foreground shrink-0 tabular-nums">
          {filtered.length} models
        </span>
        <button
          type="button"
          onClick={() => setShowFilters(!showFilters)}
          className={cn(
            "inline-flex items-center justify-center h-7 w-7 rounded hover:bg-accent transition-colors",
            showFilters && "bg-accent"
          )}
          aria-label={showFilters ? "Hide filters" : "Show filters"}
        >
          <FilterIcon className="h-3.5 w-3.5" />
        </button>
        {closeButton}
      </div>

      {/* Filter bar */}
      {showFilters && (
        <div className="flex items-center gap-1.5 px-4 py-2 border-b border-border flex-wrap">
          {/* Input modality filter */}
          <div className="relative">
            <ModalityFilterButton
              label="Input"
              selected={inputFilter}
              onClick={() =>
                setPopup(popup?.anchor === "input" ? null : { anchor: "input" })
              }
            />
            {popup?.anchor === "input" && (
              <ModalityPopup
                selected={inputFilter}
                onToggle={(mod) =>
                  setInputFilter((prev) =>
                    prev.includes(mod)
                      ? prev.filter((m) => m !== mod)
                      : [...prev, mod]
                  )
                }
                onClose={() => setPopup(null)}
              />
            )}
          </div>

          {/* Output modality filter */}
          <div className="relative">
            <ModalityFilterButton
              label="Output"
              selected={outputFilter}
              onClick={() =>
                setPopup(
                  popup?.anchor === "output" ? null : { anchor: "output" }
                )
              }
            />
            {popup?.anchor === "output" && (
              <ModalityPopup
                selected={outputFilter}
                onToggle={(mod) =>
                  setOutputFilter((prev) =>
                    prev.includes(mod)
                      ? prev.filter((m) => m !== mod)
                      : [...prev, mod]
                  )
                }
                onClose={() => setPopup(null)}
              />
            )}
          </div>

          {/* Month filter */}
          <div className="relative">
            <ModalityFilterButton
              label="Month"
              selected={monthFilter}
              onClick={() =>
                setPopup(popup?.anchor === "month" ? null : { anchor: "month" })
              }
            />
            {popup?.anchor === "month" && (
              <MonthPopup
                months={allMonths}
                selected={monthFilter}
                onToggle={(month) =>
                  setMonthFilter((prev) =>
                    prev.includes(month)
                      ? prev.filter((m) => m !== month)
                      : [...prev, month]
                  )
                }
                onClose={() => setPopup(null)}
              />
            )}
          </div>

          {/* Free toggle */}
          <button
            type="button"
            onClick={() => setFreeOnly(!freeOnly)}
            className={cn(
              "px-2.5 py-1 rounded-md text-xs transition-colors",
              freeOnly
                ? "bg-primary text-primary-foreground"
                : "bg-muted/60 text-muted-foreground hover:text-foreground"
            )}
          >
            Free
          </button>

          {/* Hide Unavailable toggle */}
          <button
            type="button"
            onClick={() => setHideUnavailable(!hideUnavailable)}
            className={cn(
              "px-2.5 py-1 rounded-md text-xs transition-colors",
              hideUnavailable
                ? "bg-primary text-primary-foreground"
                : "bg-muted/60 text-muted-foreground hover:text-foreground"
            )}
          >
            Hide Unavailable
          </button>
        </div>
      )}

      {/* Split layout — OR style: left model list, right detail */}
      <div className="flex min-h-0 flex-1">
        {/* Left: model list — scrollable */}
        <div className="w-1/2 min-w-0 border-r border-border">
          <div
            ref={listRef}
            role="listbox"
            aria-label="Suggestions"
            className="overflow-y-auto max-h-[70vh] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-border [&::-webkit-scrollbar-track]:bg-transparent [scrollbar-width:thin]"
          >
            {filtered.length === 0 ? (
              <div className="px-4 py-8 text-center text-sm text-muted-foreground">
                {debouncedSearch.trim()
                  ? `No models matching "${debouncedSearch}"`
                  : "No models available"}
              </div>
            ) : (
              <>
                {visibleSections.map((section, sectionIdx) => (
                  <div key={sectionIdx}>
                    <MonthGroup
                      label={section.month}
                      count={section.models.length}
                    />
                    {section.models.map((model) => {
                      let globalIdx = 0;
                      for (let i = 0; i < sectionIdx; i++) {
                        globalIdx += visibleSections[i].models.length;
                      }
                      globalIdx += section.models.indexOf(model);
                      return (
                        <ModelRow
                          key={model.id}
                          model={model}
                          isSelected={model.id === value}
                          pinned={pinnedSet.has(model.id)}
                          highlighted={highlightedIdx === globalIdx}
                          onSelect={() => handleSelect(model.id)}
                          onPinToggle={() => handlePinToggle(model.id)}
                          onHover={() => setHighlightedModel(model)}
                        />
                      );
                    })}
                  </div>
                ))}
              </>
            )}
          </div>
        </div>

        {/* Right: model detail panel — static, no scroll */}
        <div className="flex-1 p-4 overflow-hidden">
          {highlightedModel ? (
            <ModelDetailPanel
              model={highlightedModel}
              onSelect={() => handleSelect(highlightedModel.id)}
              browseOnly={browseOnly}
            />
          ) : (
            <div className="flex items-center justify-center h-full text-sm text-muted-foreground/60">
              Browse models to see details
            </div>
          )}
        </div>
      </div>
    </>
  );

  // ── Render ────────────────────────────────────────────

  // Inline mode: render content directly without trigger + Dialog wrapper
  if (inline) {
    return (
      <div className="w-full border border-border rounded-lg overflow-hidden">
        {renderContent()}
      </div>
    );
  }

  // Dialog mode: trigger button + Dialog wrapper
  return (
    <div className="w-full">
      {/* Trigger row */}
      <div className="flex items-center gap-2 flex-wrap min-h-10">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={disabled}
          onClick={() => {
            setOpen(true);
            onOpen?.();
          }}
          className="h-8 gap-1.5 text-sm font-normal text-muted-foreground hover:text-foreground shrink-0"
        >
          <Search className="h-3.5 w-3.5" />
          {browseOnly ? "Models available" : "Add Model"}
          <kbd className="ml-1 pointer-events-none inline-flex h-5 select-none items-center gap-0.5 rounded border border-border bg-muted px-1.5 text-[10px] font-medium text-muted-foreground">
            <span className="text-[9px]">⌘</span>J
          </kbd>
        </Button>
        {!browseOnly && selected && (
          <SelectedChip model={selected} onRemove={handleRemove} />
        )}
        {browseOnly && showCount && (
          <span className="text-xs text-muted-foreground">
            {(() => {
              const n = hideUnavailable ? routableCount : models.length;
              return `${n} model${n !== 1 ? "s" : ""} · browse only`;
            })()}
          </span>
        )}
      </div>

      {/* Pricing strip — only when selecting a model (not browse-only) */}
      {!browseOnly && selected && <ModelPricingStrip modelId={selected.id} />}

      {/* Dialog */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent
          showCloseButton={false}
          style={{ width: "90vw", maxWidth: "1280px" }}
          className="p-0 gap-0 overflow-hidden"
          onKeyDown={handleKeyDown}
        >
          <DialogHeader className="sr-only">
            <DialogTitle>
              {browseOnly ? "Models available" : "Add Model"}
            </DialogTitle>
          </DialogHeader>
          {renderContent(
            <DialogClose asChild>
              <button
                type="button"
                className="inline-flex items-center justify-center h-7 w-7 rounded hover:bg-accent transition-colors"
                aria-label="Close"
              >
                <X className="h-4 w-4" />
              </button>
            </DialogClose>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
