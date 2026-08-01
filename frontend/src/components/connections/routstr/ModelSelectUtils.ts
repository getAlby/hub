import type { RoutstrdModel } from "src/hooks/useRoutstrd";

// ─── Author Badge Map ────────────────────────────────────
// Maps upstream_provider_id → { label, ring, bg, text } colors

type BadgeStyle = {
  label: string;
  ring: string;
  bg: string;
  text: string;
};

// Colored rings + muted backgrounds (no solid fills — matches OR's clean favicon style)
const AUTHOR_BADGES: Record<string, BadgeStyle> = {
  deepseek: {
    label: "DS",
    ring: "ring-blue-500/40",
    bg: "bg-blue-500/10",
    text: "text-blue-600 dark:text-blue-400",
  },
  openai: {
    label: "OA",
    ring: "ring-emerald-500/40",
    bg: "bg-emerald-500/10",
    text: "text-emerald-600 dark:text-emerald-400",
  },
  anthropic: {
    label: "AN",
    ring: "ring-orange-500/40",
    bg: "bg-orange-500/10",
    text: "text-orange-600 dark:text-orange-400",
  },
  google: {
    label: "GO",
    ring: "ring-sky-500/40",
    bg: "bg-sky-500/10",
    text: "text-sky-600 dark:text-sky-400",
  },
  xai: {
    label: "XA",
    ring: "ring-zinc-500/40",
    bg: "bg-zinc-500/10",
    text: "text-zinc-700 dark:text-zinc-300",
  },
  qwen: {
    label: "QN",
    ring: "ring-violet-500/40",
    bg: "bg-violet-500/10",
    text: "text-violet-600 dark:text-violet-400",
  },
  mistral: {
    label: "MI",
    ring: "ring-cyan-500/40",
    bg: "bg-cyan-500/10",
    text: "text-cyan-600 dark:text-cyan-400",
  },
  meta: {
    label: "ME",
    ring: "ring-indigo-500/40",
    bg: "bg-indigo-500/10",
    text: "text-indigo-600 dark:text-indigo-400",
  },
  microsoft: {
    label: "MS",
    ring: "ring-teal-500/40",
    bg: "bg-teal-500/10",
    text: "text-teal-600 dark:text-teal-400",
  },
  "z-ai": {
    label: "ZA",
    ring: "ring-purple-500/40",
    bg: "bg-purple-500/10",
    text: "text-purple-600 dark:text-purple-400",
  },
  xiaomi: {
    label: "XM",
    ring: "ring-rose-500/40",
    bg: "bg-rose-500/10",
    text: "text-rose-600 dark:text-rose-400",
  },
  moonshotai: {
    label: "MO",
    ring: "ring-amber-500/40",
    bg: "bg-amber-500/10",
    text: "text-amber-600 dark:text-amber-400",
  },
  inclusionai: {
    label: "IN",
    ring: "ring-lime-500/40",
    bg: "bg-lime-500/10",
    text: "text-lime-600 dark:text-lime-400",
  },
  poolside: {
    label: "PS",
    ring: "ring-fuchsia-500/40",
    bg: "bg-fuchsia-500/10",
    text: "text-fuchsia-600 dark:text-fuchsia-400",
  },
  thefux: {
    label: "FX",
    ring: "ring-pink-500/40",
    bg: "bg-pink-500/10",
    text: "text-pink-600 dark:text-pink-400",
  },
  ailab: {
    label: "AL",
    ring: "ring-slate-500/40",
    bg: "bg-slate-500/10",
    text: "text-slate-600 dark:text-slate-400",
  },
  alibaba: {
    label: "AB",
    ring: "ring-red-500/40",
    bg: "bg-red-500/10",
    text: "text-red-600 dark:text-red-400",
  },
  sambanova: {
    label: "SN",
    ring: "ring-green-500/40",
    bg: "bg-green-500/10",
    text: "text-green-600 dark:text-green-400",
  },
  amazon: {
    label: "AZ",
    ring: "ring-yellow-500/40",
    bg: "bg-yellow-500/10",
    text: "text-yellow-600 dark:text-yellow-400",
  },
};

// ─── Provider Display Names ──────────────────────────────

const PROVIDER_NAMES: Record<string, string> = {
  deepseek: "DeepSeek",
  openai: "OpenAI",
  anthropic: "Anthropic",
  google: "Google",
  xai: "xAI",
  qwen: "Qwen",
  mistral: "Mistral",
  meta: "Meta",
  microsoft: "Microsoft",
  "z-ai": "Z.ai",
  xiaomi: "Xiaomi",
  moonshotai: "MoonshotAI",
  inclusionai: "InclusionAI",
  poolside: "Poolside",
  thefux: "TheFux",
  ailab: "AI Lab",
  alibaba: "Alibaba",
  sambanova: "SambaNova",
  amazon: "Amazon",
};

// ─── Modality Constants ──────────────────────────────────

const MODALITY_OPTIONS = [
  { value: "text", label: "Text" },
  { value: "image", label: "Image" },
  { value: "file", label: "File" },
  { value: "audio", label: "Audio" },
  { value: "video", label: "Video" },
] as const;

const MODALITY_ICONS: Record<string, string> = {
  text: "📝",
  image: "🖼️",
  file: "📎",
  audio: "🎤",
  video: "🎬",
};

// ─── Detection ──────────────────────────────────────────

// ─── Name-Based Fallback Detection ──────────────────────

/** Map model name prefixes to provider IDs for models
 *  with upstream_provider_id "generic" or unknown. */
const NAME_PREFIX_MAP: [string, string][] = [
  ["claude", "anthropic"],
  ["kimi", "moonshotai"],
  ["grok", "xai"],
  ["gemini", "google"],
  ["gemma", "google"],
  ["llama", "meta"],
  ["qwen", "qwen"],
  ["mistral", "mistral"],
  ["deepseek", "deepseek"],
  ["glm", "z-ai"],
];

/** Resolve the badge + human-readable author name for a model. */
export function detectAuthor(model: RoutstrdModel): {
  badge: BadgeStyle;
  author: string;
} {
  const id = (model.upstream_provider_id || "").toLowerCase().trim();
  const name = model.name || model.id || "";

  // Direct lookup for known provider IDs
  if (AUTHOR_BADGES[id]) {
    return { badge: AUTHOR_BADGES[id], author: PROVIDER_NAMES[id] || id };
  }

  // For "generic" or unknown IDs, try name-based detection
  if (id === "generic" || !AUTHOR_BADGES[id]) {
    const lowerName = name.toLowerCase();
    for (const [prefix, providerId] of NAME_PREFIX_MAP) {
      if (lowerName.startsWith(prefix)) {
        const badge = AUTHOR_BADGES[providerId];
        if (badge) {
          return { badge, author: PROVIDER_NAMES[providerId] || providerId };
        }
      }
    }
  }

  // Fallback: detect from name prefix in AUTHOR_BADGES
  for (const [key, badge] of Object.entries(AUTHOR_BADGES)) {
    if (name.toLowerCase().startsWith(key)) {
      return { badge, author: PROVIDER_NAMES[key] || key };
    }
    const providerName = PROVIDER_NAMES[key] || "";
    if (providerName && name.startsWith(providerName)) {
      return { badge, author: providerName };
    }
  }

  // Generic fallback
  return {
    badge: {
      label: (id.slice(0, 2) || "?").toUpperCase(),
      ring: "ring-gray-400/40",
      bg: "bg-gray-500/10",
      text: "text-gray-600 dark:text-gray-400",
    },
    author: id || "unknown",
  };
}

// ─── Provider Helpers ────────────────────────────────────

/** Get unique, sorted provider entries for the Provider filter. */
export function getUniqueProviders(models: RoutstrdModel[]): Array<{
  id: string;
  name: string;
  count: number;
}> {
  const counts = new Map<string, number>();
  for (const m of models) {
    const id = (m.upstream_provider_id || "unknown").toLowerCase();
    counts.set(id, (counts.get(id) || 0) + 1);
  }
  return Array.from(counts.entries())
    .map(([id, count]) => ({
      id,
      name: PROVIDER_NAMES[id] || id,
      count,
    }))
    .sort((a, b) => a.name.localeCompare(b.name));
}

/** Human-readable provider name. */
export function getProviderName(id: string): string {
  return PROVIDER_NAMES[id.toLowerCase().trim()] || id;
}

/** Strip leading provider prefix from model name.
 *  Handles both colon-separated ("xAI: Grok 4.5") and space-separated
 *  ("OpenAI GPT-5.5") prefixes, matching against ALL known providers
 *  regardless of the model's own upstream_provider_id. */
export function getModelDisplayName(model: RoutstrdModel): string {
  const name = model.name || model.id || "";
  // Build a set of all known provider prefixes (both IDs and display names)
  const allNames = new Set(
    Object.values(PROVIDER_NAMES).filter(Boolean) as string[]
  );
  Object.keys(PROVIDER_NAMES).forEach((k) => allNames.add(k));

  // Try colon-separated: "xAI: Grok 4.5" → "Grok 4.5"
  const colonIdx = name.indexOf(":");
  if (colonIdx > 0 && colonIdx < 25) {
    const prefix = name.slice(0, colonIdx).trim();
    for (const p of allNames) {
      if (prefix.toLowerCase() === p.toLowerCase()) {
        return name.slice(colonIdx + 1).trim();
      }
    }
  }

  // Try space-separated: "OpenAI GPT-5.5" → "GPT-5.5"
  // (and other multi-word prefixes)
  for (const p of allNames) {
    if (!p) {
      continue;
    }
    if (name.toUpperCase().startsWith(p.toUpperCase() + " ")) {
      return name.slice(p.length + 1).trim();
    }
  }

  return name;
}

// ─── Pricing Formatters ──────────────────────────────────

/** Format sats-per-token → sats-per-million-tokens value. */
export function formatSatsPerM(sats?: number | null): string {
  if (sats === undefined || sats === null) {
    return "—";
  }
  if (sats === 0) {
    return "0";
  }
  const perM = sats * 1_000_000;
  if (perM < 1) {
    return "<1";
  }
  if (perM >= 10_000) {
    return Math.round(perM).toLocaleString();
  }
  return perM.toFixed(2);
}

/** Check if a model is free (both prompt and completion cost 0). */
export function isFreeModel(model: RoutstrdModel): boolean {
  const sp = model.sats_pricing;
  if (!sp) {
    return false;
  }
  return (
    (sp.prompt === 0 || sp.prompt === undefined || sp.prompt === null) &&
    (sp.completion === 0 ||
      sp.completion === undefined ||
      sp.completion === null)
  );
}

/** Format a single-line pricing string for the model list. */
export function formatPriceLine(model: RoutstrdModel): string {
  if (isFreeModel(model)) {
    return "Free";
  }
  const sp = model.sats_pricing;
  return `in ${formatSatsPerM(sp?.prompt)} · out ${formatSatsPerM(sp?.completion)}`;
}

// ─── Context Formatter ───────────────────────────────────

export function formatContext(len?: number | null): string {
  if (!len || len <= 0) {
    return "?";
  }
  if (len >= 1_000_000) {
    return (len / 1_000_000).toFixed(len % 1_000_000 === 0 ? 0 : 1) + "M";
  }
  if (len >= 1_000) {
    return Math.round(len / 1_000) + "K";
  }
  return len.toLocaleString();
}

// ─── Modality Helpers ────────────────────────────────────

export function formatModalityIcons(model: RoutstrdModel): string {
  const input = model.architecture?.input_modalities || [];
  const output = model.architecture?.output_modalities || [];
  const all = [...new Set([...input, ...output])];
  return all
    .map((m) => MODALITY_ICONS[m] || "")
    .filter(Boolean)
    .join(" ");
}

/** Check if a model has a specific modality. */
export function hasModality(
  model: RoutstrdModel,
  modality: string,
  direction: "input" | "output"
): boolean {
  const arr =
    direction === "input"
      ? model.architecture?.input_modalities || []
      : model.architecture?.output_modalities || [];
  return arr.includes(modality);
}

export { MODALITY_OPTIONS, AUTHOR_BADGES, PROVIDER_NAMES };

// ─── Pin Helpers (localStorage) ──────────────────────────

const PIN_KEY = "routstr-pinned-models";

export function getPinnedModels(): string[] {
  try {
    const raw = localStorage.getItem(PIN_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function setPinnedModels(ids: string[]): void {
  try {
    localStorage.setItem(PIN_KEY, JSON.stringify(ids));
  } catch {
    // localStorage unavailable
  }
}

export function togglePin(modelId: string): string[] {
  const current = getPinnedModels();
  const next = current.includes(modelId)
    ? current.filter((id) => id !== modelId)
    : [...current, modelId];
  setPinnedModels(next);
  return next;
}

export function isPinned(modelId: string): boolean {
  return getPinnedModels().includes(modelId);
}

// ─── Month Grouping ───────────────────────────────────────

const MONTH_NAMES = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

/** Convert a unix timestamp (seconds) to "Month YYYY" label. */
export function formatMonthLabel(created: number): string {
  if (!created || created <= 0) {
    return "Unknown";
  }
  const d = new Date(created * 1000);
  return `${MONTH_NAMES[d.getMonth()]} ${d.getFullYear()}`;
}

/** Group models by creation month, newest first.
 *  Returns array of { month, models } sorted descending by date. */
export function groupModelsByMonth(
  models: RoutstrdModel[]
): Array<{ month: string; models: RoutstrdModel[] }> {
  const groups = new Map<string, RoutstrdModel[]>();
  for (const m of models) {
    const created = m.created || 0;
    const label = created > 0 ? formatMonthLabel(created) : "Unknown";
    if (!groups.has(label)) {
      groups.set(label, []);
    }
    groups.get(label)!.push(m);
  }
  return Array.from(groups.entries())
    .map(([month, mods]) => ({ month, models: mods }))
    .sort((a, b) => {
      const da = new Date(a.month).getTime();
      const db = new Date(b.month).getTime();
      return db - da || a.month.localeCompare(b.month);
    });
}

// ─── Model Count Helpers ─────────────────────────────────

export function getEnabledCount(models: RoutstrdModel[]): number {
  return models.filter((m) => m.enabled !== false).length;
}

/** Get all unique author names from models. */
export function getModelAuthors(models: RoutstrdModel[]): string[] {
  const set = new Set<string>();
  for (const m of models) {
    const { author } = detectAuthor(m);
    if (author && author !== "unknown") {
      set.add(author);
    }
  }
  return [...set].sort((a, b) => a.localeCompare(b));
}
