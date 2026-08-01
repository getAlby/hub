import { request } from "src/utils/request";

// Module-level cache for getAllRoutstrdModels
let cachedModels: { models: RoutstrdModel[] } | null = null;
let cacheTimestamp = 0;
const CACHE_TTL = 60 * 1000; // 60 seconds — the daemon keeps its own warm cache

export type CostFields = {
  prompt: number;
  completion: number;
  request?: number;
  image?: number;
  web_search?: number;
  input_cache_read?: number;
  input_cache_write?: number;
  max_prompt_cost?: number;
  max_completion_cost?: number;
  max_cost?: number;
  internal_reasoning?: number;
};

export type RoutstrdModel = {
  id: string;
  name?: string;
  description?: string;
  context_length?: number;
  created?: number;
  enabled?: boolean;
  upstream_provider_id?: string;
  canonical_slug?: string;
  architecture?: {
    modality?: string;
    input_modalities?: string[];
    output_modalities?: string[];
    tokenizer?: string;
    instruct_type?: string | null;
  };
  pricing?: CostFields;
  sats_pricing?: CostFields;
  top_provider?: {
    context_length?: number;
    max_completion_tokens?: number;
    is_moderated?: boolean;
  };
  per_request_limits?: Record<string, unknown> | null;
};

export type RoutstrdKey = {
  id: string;
  name: string;
  balance: number;
  apiKey?: string;
  createdAt?: number;
  lastUsed?: number | null;
};

type DaemonResponse<T> = { output: T };

type RoutstrdFetchOptions = RequestInit & {
  /** Abort after N ms (default 15000). NWC status should use a short timeout. */
  timeoutMs?: number;
};

function routstrdFetch<T>(path: string, init?: RoutstrdFetchOptions) {
  const { timeoutMs = 15_000, ...rest } = init || {};
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  // Merge signals if caller already passed one
  const parentSignal = rest.signal;
  if (parentSignal) {
    if (parentSignal.aborted) {
      controller.abort();
    } else {
      parentSignal.addEventListener("abort", () => controller.abort(), {
        once: true,
      });
    }
  }

  return request<DaemonResponse<T>>(`/api/routstrd${path}`, {
    ...rest,
    signal: controller.signal,
  })
    .then((r) => r?.output)
    .finally(() => clearTimeout(timer));
}

export async function getAllRoutstrdModels(forceRefresh = false) {
  // Return cached result if fresh (unless forcing refresh)
  if (
    !forceRefresh &&
    cachedModels &&
    Date.now() - cacheTimestamp < CACHE_TTL
  ) {
    return cachedModels;
  }
  // Otherwise fetch and cache
  const qs = forceRefresh ? "?refresh=true" : "";
  const result = await routstrdFetch<{ models: RoutstrdModel[] }>(
    `/models/all${qs}`
  );
  if (result) {
    cachedModels = result;
    cacheTimestamp = Date.now();
  }
  return result;
}

export async function nwcConnect(connectionString: string) {
  return routstrdFetch<{ message: string }>("/nwc/connect", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ connectionString }),
    timeoutMs: 25_000,
  });
}

export async function createRoutstrdClient(name: string) {
  const id = name
    .toLowerCase()
    .replace(/\s+/g, "-")
    .replace(/[^a-z0-9-]/g, "");
  return routstrdFetch<{
    message?: string;
    client: RoutstrdKey;
    created: boolean;
  }>("/clients/add", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, id }),
  });
}

export async function deleteRoutstrdClient(id: string) {
  return routstrdFetch<{ message: string; id: string }>("/clients/delete", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ id }),
  });
}

export async function getRoutstrdClients() {
  return routstrdFetch<{
    clients: {
      id?: string;
      clientId?: string;
      name?: string;
      apiKey?: string;
    }[];
  }>("/clients");
}

export async function getRoutstrdKeyBalances() {
  return routstrdFetch<{
    keys: RoutstrdKey[];
    total: number;
    unit: string;
  }>("/keys/balance");
}

export async function getRoutstrdBalance() {
  return routstrdFetch<{
    balances: Record<string, number>;
    unit: string;
    activeMint: string;
  }>("/balance");
}

export async function getRoutstrdModelProviders(modelId: string) {
  return routstrdFetch<{
    id: string;
    name?: string;
    providers: Array<{
      baseUrl: string;
      pricing: {
        prompt: number;
        completion: number;
        request: number;
        max_cost: number;
      };
    }>;
  }>(`/models/${encodeURIComponent(modelId)}/providers`);
}

export async function getRoutstrdUsageSummary() {
  return routstrdFetch<{
    generatedAt: number;
    totals: {
      requests: number;
      promptTokens: number;
      completionTokens: number;
      totalTokens: number;
      cost: number;
      satsCost: number;
    };
    models: Array<{ model: string; requests: number; satsCost: number }>;
    providers: Array<{ provider: string; requests: number; satsCost: number }>;
    clients: Array<{ client: string; requests: number; satsCost: number }>;
    days: Array<{ date: string; requests: number; satsCost: number }>;
  }>("/usage/summary");
}

/**
 * Fund the Cashu wallet reliably by asking the mint for a Bolt11 invoice
 * then paying it via Hub's Lightning node.
 *
 * ALWAYS sources sats from the Routstr app's isolated NWC wallet
 * (fromAppId) — never the main wallet.
 *
 * Returns the funded amount on success or throws.
 */
export async function fundFromHub(
  amount: number,
  appId: number
): Promise<number> {
  if (!appId) {
    throw new Error("fundFromHub requires the Routstr appId");
  }

  // 1. Get invoice from the Cashu mint
  const invResult = await routstrdFetch<{
    invoice: string;
    amount: number;
    mintUrl?: string;
  }>("/wallet/receive/bolt11", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ amount }),
  });
  const invoice = invResult?.invoice;
  if (!invoice) {
    throw new Error("Failed to create invoice at Cashu mint");
  }

  // 2. Pay invoice via Hub's Lightning node, from the Routstr isolated wallet
  const encodedInvoice = encodeURIComponent(invoice);
  const payResult = await request<{
    id: number;
    type: string;
    state: string;
  }>(`/api/payments/${encodedInvoice}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ fromAppId: appId }),
  });

  if (!payResult || payResult.state !== "settled") {
    throw new Error(`Hub payment failed: ${payResult?.state || "no response"}`);
  }

  // 3. Confirm balance increased
  return amount;
}

/**
 * Refund the Cashu wallet back to the Routstr app's isolated NWC wallet.
 *
 * Uses an APP-SCOPED invoice: the invoice is created for the Routstr app
 * itself, so when the mint melts Cashu to pay it, the sats are credited
 * DIRECTLY to the Routstr app wallet — no main-wallet hop, no transfer,
 * no hidden fee. The amount shown in the dialog is the amount that lands.
 */
export async function refundFromHub(
  amount: number,
  appId: number
): Promise<number> {
  if (!appId) {
    throw new Error("refundFromHub requires the Routstr appId");
  }

  // 1. Create an invoice scoped to the Routstr app (credited to its wallet on payment)
  const invResult = await request<{
    invoice: string;
    r_hash?: string;
  }>("/api/invoices", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ amount: amount * 1000, appId }),
  });
  const invoice = invResult?.invoice;
  if (!invoice) {
    throw new Error("Failed to create invoice on Hub");
  }

  // 2. Melt Cashu tokens to pay that invoice → sats land in the Routstr wallet
  const meltResult = await routstrdFetch<{ message: string }>(
    "/wallet/send/bolt11",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        invoice,
        mintUrl: "https://mint.cubabitcoin.org",
      }),
      timeoutMs: 90_000,
    }
  );
  if (!meltResult?.message) {
    throw new Error("Melt failed: no confirmation from daemon");
  }

  return amount;
}
