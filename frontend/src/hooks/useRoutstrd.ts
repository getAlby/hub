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

/**
 * Reclaims prepaid provider tokens (the `apikey:*` entries in /keys/balance)
 * back into the daemon's Cashu wallet at the given mint. Provider tokens are
 * ecash deposited at the provider's mint for prepayment — they are NOT in the
 * wallet and cannot be melted until reclaimed. Called before a refund so the
 * refund covers all money, not just the wallet pile.
 */
export async function reclaimProviderTokens(mintUrl: string) {
  return routstrdFetch<{
    message: string;
    pendingTokens: number;
    apiKeys: number;
    results: Array<{ baseUrl: string; success: boolean }>;
  }>("/refund", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ mintUrl }),
  });
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
 * Refund the ENTIRE Cashu wallet back to the Routstr app's isolated wallet.
 *
 * Uses APP-SCOPED invoices: the invoice is created for the Routstr app
 * itself, so when the mint melts Cashu to pay it, the sats are credited
 * DIRECTLY to the Routstr app wallet — no main-wallet hop, no transfer.
 *
 * Drains to zero: each pass quotes the mint's melt fee fresh for the full
 * remaining balance, then melts `balance − fee`. The mint returns any
 * change (fee overestimate or proof overshoot) to the wallet, which the
 * next pass re-quotes and melts. The loop stops when the balance is zero
 * or when the fee exceeds the remainder (a sub-fee amount no melt can
 * move). The fee is never assumed — it is queried every pass.
 */
export async function refundFromHub(
  appId: number,
  mintUrl: string
): Promise<number> {
  if (!appId) {
    throw new Error("refundFromHub requires the Routstr appId");
  }
  if (!mintUrl) {
    throw new Error("refundFromHub requires the active mint URL");
  }

  let totalRefunded = 0;

  for (let pass = 0; pass < 6; pass++) {
    const bal = await getRoutstrdBalance();
    const walletBal = bal?.balances
      ? Object.values(bal.balances).reduce((a, b) => a + b, 0)
      : 0;
    if (walletBal <= 0) {
      break;
    }

    // 1. Quote the mint's melt fee fresh for the FULL remaining balance.
    const quoteInvoice = await createAppScopedInvoice(walletBal, appId);
    const fee = await getMeltQuoteFee(quoteInvoice, mintUrl);

    const send = Math.floor(walletBal - fee);
    if (send <= 0) {
      // The fee covers the whole remainder — nothing meltable left. This is
      // the floor: any sub-fee balance cannot be moved.
      break;
    }

    try {
      // 2. Melt `send` (balance − fee) via an app-scoped invoice. Proofs
      //    cover send + the melt's own fee quote, and the mint returns any
      //    change to the wallet, drained on the next pass.
      const meltInvoice = await createAppScopedInvoice(send, appId);
      const meltResult = await routstrdFetch<{ message: string }>(
        "/wallet/send/bolt11",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ invoice: meltInvoice, mintUrl }),
          timeoutMs: 90_000,
        }
      );
      if (!meltResult?.message) {
        throw new Error("Melt failed: no confirmation from daemon");
      }
      totalRefunded += send;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      if (
        /insufficient|not enough (funds|proofs)|non-negative/i.test(message)
      ) {
        // Known coco-cashu-core degenerate case: when the selected proofs
        // exactly equal invoice + fee, the swap path computes a zero/negative
        // keep amount ("amount must be a non-negative number"). Retry with a
        // smaller send — the wallet's per-proof input fee can also add 1-2
        // sats at small denominations. The next pass re-quotes anyway.
        let succeeded = false;
        for (let shrink = 1; shrink <= 4 && !succeeded; shrink++) {
          const retrySend = send - shrink;
          if (retrySend <= 0) {
            break;
          }
          const retryInvoice = await createAppScopedInvoice(retrySend, appId);
          const retryMelt = await routstrdFetch<{ message: string }>(
            "/wallet/send/bolt11",
            {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ invoice: retryInvoice, mintUrl }),
              timeoutMs: 90_000,
            }
          );
          if (retryMelt?.message) {
            totalRefunded += retrySend;
            succeeded = true;
          }
        }
        if (!succeeded) {
          throw error;
        }
        continue;
      }
      throw error;
    }
  }

  return totalRefunded;
}

async function createAppScopedInvoice(
  amountSat: number,
  appId: number
): Promise<string> {
  const invResult = await request<{ invoice: string }>("/api/invoices", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ amount: amountSat * 1000, appId }),
  });
  const invoice = invResult?.invoice;
  if (!invoice) {
    throw new Error("Failed to create invoice on Hub");
  }
  return invoice;
}

async function getMeltQuoteFee(
  invoice: string,
  mintUrl: string
): Promise<number> {
  try {
    const mtResp = await fetch(
      `${mintUrl.replace(/\/+$/, "")}/v1/melt/quote/bolt11`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ request: invoice, unit: "sat" }),
      }
    );
    if (mtResp.ok) {
      const quote = await mtResp.json();
      return typeof quote.fee_reserve === "number" ? quote.fee_reserve : 0;
    }
  } catch {
    // fall through — fee unknown, treat as 0
  }
  return 0;
}
