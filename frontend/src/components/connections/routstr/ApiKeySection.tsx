import { useState, useEffect, useCallback } from "react";
import { PlusIcon, RefreshCwIcon } from "lucide-react";
import { toast } from "sonner";
import { DeleteKeyDialog } from "src/components/connections/routstr/DeleteKeyDialog";
import { RefundDialog } from "src/components/connections/routstr/RefundDialog";
import { TopUpDialog } from "src/components/connections/routstr/TopUpDialog";
import ModelSelect from "src/components/connections/routstr/ModelSelect";
import { RoutstrConnectionDetails } from "src/components/connections/routstr/RoutstrConnectionDetails";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { request } from "src/utils/request";
import { handleRequestError } from "src/utils/handleRequestError";
import {
  RoutstrdModel,
  createRoutstrdClient,
  getAllRoutstrdModels,
  getRoutstrdClients,
  getRoutstrdKeyBalances,
  getRoutstrdUsageSummary,
} from "src/hooks/useRoutstrd";
import type { App } from "src/types";

const ROUTSTR_APP_ID = "routstr";

type RoutstrMetadata = {
  apiKey?: string;
  modelId?: string;
  clientId?: string;
  balance?: number;
  provider?: string;
  autoRefill?: {
    enabled: boolean;
    threshold: number;
    amount: number;
    cooldownMs?: number;
  };
};

type Props = {
  app: App;
  onMetadataUpdate: () => void;
};

export function RoutstrApiKeySection({ app, onMetadataUpdate }: Props) {
  // No hooks here — safe to early-return. The inner component holds ALL
  // hooks so a guard flip between renders can never change the hook count
  // (React error #300 — hit 2026-07-31 when app metadata lost
  // app_store_app_id mid-session and the conditional return flipped).
  if (app.metadata?.app_store_app_id !== ROUTSTR_APP_ID) {
    return null;
  }
  return (
    <RoutstrApiKeySectionInner app={app} onMetadataUpdate={onMetadataUpdate} />
  );
}

function RoutstrApiKeySectionInner({ app, onMetadataUpdate }: Props) {
  const [showCreate, setShowCreate] = useState(false);
  const [showTopUp, setShowTopUp] = useState(false);
  const [showRefund, setShowRefund] = useState(false);
  const [showDelete, setShowDelete] = useState(false);
  const [balanceRefreshing, setBalanceRefreshing] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [daemonStatus, setDaemonStatus] = useState<{
    service_running: boolean;
    routstrd_healthy: boolean;
    cocod_healthy: boolean;
    last_error?: string;
  } | null>(null);
  const [usage, setUsage] = useState<{
    totals: { requests: number; satsCost: number };
    models: Array<{ model: string; requests: number; satsCost: number }>;
    clients: Array<{ client: string; requests: number; satsCost: number }>;
  } | null>(null);
  const [liveBalance, setLiveBalance] = useState<number | null>(null);
  const [models, setModels] = useState<RoutstrdModel[]>([]);
  const [autoRefill, setAutoRefill] = useState<
    NonNullable<RoutstrMetadata["autoRefill"]>
  >(
    () =>
      (app.metadata?.routstr as RoutstrMetadata | undefined)?.autoRefill ?? {
        enabled: false,
        threshold: 500,
        amount: 1000,
      }
  );
  const [savingAutoRefill, setSavingAutoRefill] = useState(false);
  const [autoRefillStatus, setAutoRefillStatus] = useState<{
    enabled: boolean;
    appId: number;
    threshold: number;
    amount: number;
    cooldownMs: number;
    poolBalanceSat: number;
    routstrWalletSat: number;
    lastCheckAt?: string;
    lastRefillAt?: string;
    lastRefillAmount?: number;
    lastError?: string;
    routstrdHealthy: boolean;
    cocodHealthy: boolean;
  } | null>(null);

  const routstrMeta: RoutstrMetadata = (app.metadata?.routstr ||
    {}) as RoutstrMetadata;
  const hasKey = !!routstrMeta.apiKey;
  const displayBalance =
    liveBalance !== null ? liveBalance : (routstrMeta.balance ?? null);

  useEffect(() => {
    if (!hasKey) {
      return;
    }
    let cancelled = false;

    (async () => {
      try {
        const summary = await getRoutstrdUsageSummary();
        if (!cancelled && summary && summary.totals) {
          setUsage(summary);
        }
      } catch {
        // Non-critical
      }
    })();

    (async () => {
      try {
        const res = await fetch("/api/health");
        if (!res.ok) {
          return;
        }
        const health = await res.json();
        if (!cancelled && health?.alarms) {
          const routstrdAlarm = health.alarms.find(
            (a: { kind: string }) => a.kind === "routstrd_offline"
          );
          if (routstrdAlarm?.rawDetails) {
            setDaemonStatus({
              service_running: true,
              routstrd_healthy:
                routstrdAlarm.rawDetails.routstrd_healthy ?? false,
              cocod_healthy: routstrdAlarm.rawDetails.cocod_healthy ?? false,
              last_error: routstrdAlarm.rawDetails.last_error,
            });
          } else {
            setDaemonStatus({
              service_running: true,
              routstrd_healthy: true,
              cocod_healthy: true,
            });
          }
        }
      } catch {
        // Non-critical
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [hasKey]);

  // Auto top-up status: fetch on mount and keep it live while the card is
  // shown (pool balance and last refill move as requests drain the pool).
  const refreshAutoRefillStatus = useCallback(async () => {
    try {
      const status = await request<NonNullable<typeof autoRefillStatus>>(
        "/api/routstrd/autorefill/status"
      );
      if (status) {
        setAutoRefillStatus(status);
      }
    } catch {
      // Non-critical
    }
  }, []);
  useEffect(() => {
    if (!hasKey) {
      return;
    }
    refreshAutoRefillStatus();
    const interval = setInterval(refreshAutoRefillStatus, 30_000);
    return () => clearInterval(interval);
  }, [hasKey, refreshAutoRefillStatus]);

  // Load models on mount + silently refresh when dialog opens
  const refreshModels = useCallback(async () => {
    try {
      const result = await getAllRoutstrdModels();
      if (result?.models) {
        setModels(result.models);
      }
    } catch {
      // keep existing cache
    }
  }, []);
  useEffect(() => {
    if (!hasKey) {
      return;
    }
    refreshModels();
  }, [hasKey, refreshModels]);
  const onModelsOpen = useCallback(() => {
    // Silently refresh models in background — user sees cached list instantly
    refreshModels();
  }, [refreshModels]);

  const refreshLiveBalance = useCallback(
    async (opts?: { silent?: boolean; persist?: boolean }) => {
      const silent = opts?.silent ?? true;
      const persist = opts?.persist ?? true;
      try {
        const result = await getRoutstrdKeyBalances();
        const total =
          typeof result?.total === "number"
            ? result.total
            : (result?.keys || []).reduce(
                (sum, k) => sum + (k.balance || 0),
                0
              );
        setLiveBalance(total);
        // Read-modify-write against the SERVER, never the closure. The
        // self-heal and balance persist below write metadata, and a stale
        // snapshot would clobber a key created while this balance fetch was
        // in flight (hit 2026-07-31 — the mount-time persist wiped freshly
        // created key metadata, leaving orphan daemon clients).
        const freshApp = (await request(`/api/v2/apps/${app.id}`)) as {
          metadata?: { routstr?: RoutstrMetadata };
        };
        const freshMeta = (freshApp?.metadata || {}) as Record<string, unknown>;
        const freshRoutstr = (freshMeta.routstr || {}) as RoutstrMetadata;
        const freshHasKey = !!freshRoutstr.apiKey;
        // Self-heal ghost keys: metadata points at a client the daemon no
        // longer has (deleted directly, daemon DB restored, wizard abandoned).
        // Without this the page shows a dead "Active" key whose Delete button
        // 404s and never clears. Check the daemon's /clients list (NOT
        // /keys/balance — that returns wallet + provider-token entries, never
        // client ids, which would make EVERY key look like a ghost).
        let daemonIds: (string | undefined)[] = [];
        try {
          const clientsResult = await getRoutstrdClients();
          daemonIds = (clientsResult?.clients || []).map(
            (c) => c.id ?? c.clientId
          );
        } catch {
          /* daemon unreachable — skip the self-heal this cycle */
        }
        if (
          freshRoutstr.clientId &&
          daemonIds.length > 0 &&
          !daemonIds.includes(freshRoutstr.clientId)
        ) {
          try {
            await request(`/api/apps/${app.appPubkey}`, {
              method: "PATCH",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({
                metadata: {
                  app_store_app_id: ROUTSTR_APP_ID,
                  ...freshMeta,
                  routstr: {
                    ...freshRoutstr,
                    apiKey: undefined,
                    clientId: undefined,
                    balance: undefined,
                    modelId: undefined,
                  },
                },
              }),
            });
            onMetadataUpdate();
          } catch {
            /* next mount retries */
          }
          return total;
        }
        if (persist && freshHasKey) {
          const newMetadata = {
            app_store_app_id: ROUTSTR_APP_ID,
            ...freshMeta,
            routstr: { ...freshRoutstr, balance: total },
          };
          try {
            await request(`/api/apps/${app.appPubkey}`, {
              method: "PATCH",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ metadata: newMetadata }),
            });
            if (!silent) {
              onMetadataUpdate();
            }
          } catch {
            /* display still uses liveBalance */
          }
        }
        return total;
      } catch (e) {
        if (!silent) {
          toast.error("Failed to refresh balance");
        }
        throw e;
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [hasKey, app.appPubkey, app.metadata, onMetadataUpdate]
  );

  // Live balance on mount + tab focus
  useEffect(() => {
    if (!hasKey) {
      setLiveBalance(null);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        await refreshLiveBalance({ silent: true, persist: true });
      } catch {
        if (!cancelled) {
          // refresh already ran; nothing to do
        }
      }
    })();

    let lastFocusFetch = 0;
    const onVisible = () => {
      if (document.visibilityState !== "visible") {
        return;
      }
      const now = Date.now();
      if (now - lastFocusFetch < 2000) {
        return;
      }
      lastFocusFetch = now;
      refreshLiveBalance({ silent: true, persist: true }).catch(() => {});
    };
    document.addEventListener("visibilitychange", onVisible);
    window.addEventListener("focus", onVisible);
    return () => {
      cancelled = true;
      document.removeEventListener("visibilitychange", onVisible);
      window.removeEventListener("focus", onVisible);
    };
  }, [hasKey, refreshLiveBalance]);

  const loadModels = useCallback(async () => {
    try {
      const result = await getAllRoutstrdModels();
      if (result?.models) {
        // Pre-fetched for CreateKeyDialog; no UI display in this section
      }
    } catch (e) {
      console.error("Failed to load models", e);
    }
  }, []);

  useEffect(() => {
    if (showCreate) {
      loadModels();
    }
  }, [showCreate, loadModels]);

  const updateAppMetadata = async (
    updates: Omit<Partial<RoutstrMetadata>, "autoRefill"> & {
      autoRefill?: Partial<NonNullable<RoutstrMetadata["autoRefill"]>>;
    }
  ) => {
    try {
      // Read-modify-write against the SERVER, never the closure: a key
      // created while this call was queued must not be clobbered by a
      // stale snapshot (hit 2026-07-31 — create-key metadata was wiped by
      // the in-flight balance persist, leaving orphan daemon clients).
      const freshApp = (await request(`/api/v2/apps/${app.id}`)) as {
        metadata?: { routstr?: RoutstrMetadata };
      };
      const freshMeta = (freshApp?.metadata || {}) as Record<string, unknown>;
      const freshRoutstr = (freshMeta.routstr || {}) as RoutstrMetadata;
      // autoRefill merges DEEP: `enabled` is owned by the Start/Stop API and
      // must survive this write unchanged. A shallow spread of `updates`
      // would drop it (undefined), which the server reads as disabled and
      // silently stops a running auto top-up (hit 2026-08-01: blur-save
      // raced Start and the PATCH landed after the start POST, stopping it).
      const mergedAutoRefill = {
        ...(freshRoutstr.autoRefill || {}),
        ...(updates.autoRefill || {}),
      };
      // Always re-assert the app-store id — stale writebacks wiped it from
      // some apps' metadata (2026-07-31), which hid the Routstr section.
      const newMetadata = {
        app_store_app_id: ROUTSTR_APP_ID,
        ...freshMeta,
        routstr: {
          ...freshRoutstr,
          ...updates,
          autoRefill: mergedAutoRefill,
        },
      };
      await request(`/api/apps/${app.appPubkey}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ metadata: newMetadata }),
      });
      onMetadataUpdate();
    } catch (error) {
      console.error("Failed to update app metadata", error);
      toast.error("Failed to save key details");
    }
  };

  const handleSaveAutoRefill = async (
    next: NonNullable<RoutstrMetadata["autoRefill"]>
  ) => {
    // `enabled` is owned by the Start/Stop API (server-side). This handler
    // runs on input blur (including blur-on-unmount when the inputs disappear
    // after a stop) — writing `enabled` here resurrects a stale value after a
    // server-side stop, or stops a just-started loop when the blur-save races
    // the Start click (hit 2026-08-01: PATCH landed after the start POST).
    // updateAppMetadata deep-merges autoRefill, so the server's current
    // enabled survives; only threshold/amount/cooldown are written.
    const { enabled: _serverOwned, ...values } = next;
    setAutoRefill(next);
    try {
      await updateAppMetadata({
        autoRefill: {
          ...values,
          cooldownMs: values.cooldownMs ?? 5 * 60 * 1000,
        },
      });
      toast.success(
        _serverOwned
          ? `Auto top-up on: refills ${values.amount} sats below ${values.threshold} sats`
          : "Auto top-up off"
      );
    } finally {
      // No button loading here: this runs on input blur (including when the
      // user clicks Start/Stop), and toggling `savingAutoRefill` would
      // disable the Start button for the blur-save's duration — swallowing
      // the click that triggered the blur (hit 2026-08-01).
    }
  };

  const handleStartAutoRefill = async () => {
    setSavingAutoRefill(true);
    try {
      // Send the typed values with the start: the server persists them
      // atomically with enabled=true, so what the user entered is what the
      // loop honors (no blur-save race).
      const status = await request<NonNullable<typeof autoRefillStatus>>(
        "/api/routstrd/autorefill/start",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            threshold: autoRefill.threshold,
            amount: autoRefill.amount,
          }),
        }
      );
      if (status) {
        setAutoRefillStatus(status);
        setAutoRefill((prev) => ({
          ...prev,
          enabled: status.enabled,
          threshold: status.threshold,
          amount: status.amount,
        }));
      }
      onMetadataUpdate();
      toast.success("Auto top-up started");
    } catch (error) {
      handleRequestError("Failed to start auto top-up", error);
    } finally {
      setSavingAutoRefill(false);
    }
  };

  const handleStopAutoRefill = async () => {
    setSavingAutoRefill(true);
    try {
      const status = await request<NonNullable<typeof autoRefillStatus>>(
        "/api/routstrd/autorefill/stop",
        { method: "POST" }
      );
      if (status) {
        setAutoRefillStatus(status);
        setAutoRefill((prev) => ({ ...prev, enabled: status.enabled }));
      }
      onMetadataUpdate();
      toast.success("Auto top-up stopped");
    } catch (error) {
      handleRequestError("Failed to stop auto top-up", error);
    } finally {
      setSavingAutoRefill(false);
    }
  };

  const handleCreateKey = async () => {
    setIsCreating(true);
    try {
      // Key creation is free and independent of funding — create it first,
      // then the user funds via Top Up (or auto-refill). Previously this
      // funded first, which failed on an empty Routstr wallet or when the
      // mint's Lightning backend is unavailable (chicken-and-egg: you could
      // never create a key to top up).
      const clientName = `routstr-app-${app.id}-${Date.now()}`;
      const clientResult = await createRoutstrdClient(clientName);
      const apiKey = clientResult?.client?.apiKey;
      const clientId = clientResult?.client?.id || clientName;
      if (!apiKey) {
        throw new Error("Failed to create API key");
      }
      await updateAppMetadata({
        apiKey,
        clientId,
        balance: 0,
        modelId: undefined,
      });
      setShowCreate(false);
      toast.success("API key created!");
    } catch (error) {
      handleRequestError("Failed to create API key", error);
    } finally {
      setIsCreating(false);
    }
  };

  const handleTopUp = async (_amount: number) => {
    try {
      const total = await refreshLiveBalance({ silent: true, persist: true });
      if (typeof total === "number") {
        await updateAppMetadata({ balance: total });
        onMetadataUpdate();
      }
    } catch {
      /* ignore */
    }
  };

  const handleRefundComplete = async () => {
    try {
      const total = await refreshLiveBalance({ silent: true, persist: true });
      await updateAppMetadata({
        balance: typeof total === "number" ? total : 0,
      });
      onMetadataUpdate();
    } catch {
      await updateAppMetadata({ balance: 0 });
    }
  };

  const handleDeleted = () => {
    updateAppMetadata({
      apiKey: undefined,
      modelId: undefined,
      clientId: undefined,
      balance: undefined,
      provider: undefined,
    });
  };

  const handleRefreshBalance = async () => {
    setBalanceRefreshing(true);
    try {
      const totalBalance = await refreshLiveBalance({
        silent: false,
        persist: true,
      });
      if (typeof totalBalance === "number") {
        await updateAppMetadata({ balance: totalBalance });
        onMetadataUpdate();
        toast.success(`Balance refreshed: ${totalBalance} sats`);
      }
    } catch {
      toast.error("Failed to refresh balance");
    } finally {
      setBalanceRefreshing(false);
    }
  };

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">Routstr API Key</CardTitle>
            <div className="flex items-center gap-2">
              {daemonStatus?.routstrd_healthy &&
                daemonStatus?.cocod_healthy && (
                  <Badge variant="positive">Daemon</Badge>
                )}
              {daemonStatus &&
                !(
                  daemonStatus.routstrd_healthy && daemonStatus.cocod_healthy
                ) && <Badge variant="warning">Daemon</Badge>}
              {hasKey && <Badge variant="positive">Active</Badge>}
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {hasKey ? (
            <>
              {/* API Balance row */}
              <div className="flex items-end justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">
                      API Balance
                    </span>
                    <button
                      type="button"
                      onClick={handleRefreshBalance}
                      disabled={balanceRefreshing}
                      className="inline-flex items-center gap-1 text-xs text-muted-foreground/60 hover:text-foreground transition-colors disabled:opacity-50"
                    >
                      <RefreshCwIcon
                        className={`h-2.5 w-2.5 ${balanceRefreshing ? "animate-spin" : ""}`}
                      />
                      refresh
                    </button>
                  </div>
                  <p className="text-2xl font-bold tabular-nums">
                    {displayBalance !== null ? displayBalance : "?"} sats
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <ModelSelect
                    models={models}
                    browseOnly
                    showCount={false}
                    onOpen={onModelsOpen}
                  />
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => setShowTopUp(true)}
                  >
                    <PlusIcon className="h-3.5 w-3.5 mr-1" />
                    Top Up
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setShowRefund(true)}
                  >
                    Refund
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() => setShowDelete(true)}
                  >
                    Delete Key
                  </Button>
                </div>
              </div>

              {/* Key + endpoints */}
              <RoutstrConnectionDetails
                apiKey={routstrMeta.apiKey!}
                showHubUrl
              />

              {/* Usage summary: scoped to THIS app's daemon client, not the
                  daemon-wide totals (which span every key ever created). */}
              {(() => {
                const myUsage = usage?.clients?.find(
                  (c) => c.client === routstrMeta.clientId
                );
                const myRequests = myUsage?.requests ?? 0;
                if (!myUsage || myRequests === 0) {
                  return null;
                }
                return (
                  <div className="flex gap-4 text-sm text-muted-foreground">
                    <span>
                      {myRequests} {myRequests === 1 ? "request" : "requests"}
                    </span>
                    <span>{myUsage.satsCost.toFixed(2)} sats spent</span>
                  </div>
                );
              })()}

              {/* Auto top-up */}
              <div className="rounded-lg border border-border p-3 space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">Auto top-up</p>
                    <p className="text-xs text-muted-foreground">
                      When the API key balance drops, top up automatically from
                      your Routstr wallet.
                    </p>
                  </div>
                  {autoRefill.enabled ? (
                    <LoadingButton
                      loading={savingAutoRefill}
                      variant="outline"
                      size="sm"
                      onClick={handleStopAutoRefill}
                    >
                      Stop
                    </LoadingButton>
                  ) : (
                    <LoadingButton
                      loading={savingAutoRefill}
                      size="sm"
                      onClick={handleStartAutoRefill}
                    >
                      Start
                    </LoadingButton>
                  )}
                </div>
                {autoRefillStatus && (
                  <div className="text-xs text-muted-foreground">
                    {autoRefill.enabled ? (
                      <span>
                        Pool balance {autoRefillStatus.poolBalanceSat} sats
                        {autoRefillStatus.lastRefillAt &&
                          ` · last refill ${
                            autoRefillStatus.lastRefillAmount
                              ? `${autoRefillStatus.lastRefillAmount} sats at `
                              : ""
                          }${new Date(
                            autoRefillStatus.lastRefillAt
                          ).toLocaleTimeString()}`}
                      </span>
                    ) : (
                      <span>Stopped</span>
                    )}
                    {autoRefillStatus.lastError && (
                      <span className="text-amber-600 dark:text-amber-400">
                        {" "}
                        · {autoRefillStatus.lastError}
                      </span>
                    )}
                  </div>
                )}
                {autoRefill.enabled &&
                  (app.balanceMsat ?? 0) / 1000 < autoRefill.amount && (
                    <p className="text-xs text-amber-600 dark:text-amber-400">
                      ⚠️ Auto top-up is on, but your Routstr wallet has{" "}
                      {Math.floor((app.balanceMsat ?? 0) / 1000)} sats — top it
                      up to keep the Cashu wallet funded automatically.
                    </p>
                  )}
                {autoRefill.enabled && autoRefill.amount < 100 && (
                  <p className="text-xs text-amber-600 dark:text-amber-400">
                    ⚠️ Each refill pays a network fee. Small amounts can go
                    mostly to fees.
                  </p>
                )}
                {/* Values are editable before AND while running: type your
                    threshold/amount, press Start, and the loop honors exactly
                    those values. */}
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>Top up</span>
                  <Input
                    type="number"
                    min={1}
                    value={autoRefill.amount}
                    onChange={(e) =>
                      setAutoRefill({
                        ...autoRefill,
                        amount: Math.max(1, Number(e.target.value) || 1),
                      })
                    }
                    onBlur={() => handleSaveAutoRefill(autoRefill)}
                    className="h-7 w-24 tabular-nums"
                  />
                  <span>sats when balance &lt;</span>
                  <Input
                    type="number"
                    min={1}
                    value={autoRefill.threshold}
                    onChange={(e) =>
                      setAutoRefill({
                        ...autoRefill,
                        threshold: Math.max(1, Number(e.target.value) || 1),
                      })
                    }
                    onBlur={() => handleSaveAutoRefill(autoRefill)}
                    className="h-7 w-24 tabular-nums"
                  />
                  <span>sats</span>
                </div>
              </div>
            </>
          ) : (
            <div className="flex gap-2">
              <LoadingButton loading={isCreating} onClick={handleCreateKey}>
                <PlusIcon className="h-3.5 w-3.5 mr-1" />
                Create API Key
              </LoadingButton>
            </div>
          )}
        </CardContent>
      </Card>

      <TopUpDialog
        open={showTopUp}
        onOpenChange={setShowTopUp}
        onTopUp={handleTopUp}
        appId={app.id}
      />
      <RefundDialog
        open={showRefund}
        onOpenChange={setShowRefund}
        onRefundComplete={handleRefundComplete}
        appId={app.id}
      />
      <DeleteKeyDialog
        open={showDelete}
        onOpenChange={setShowDelete}
        clientId={routstrMeta.clientId || `routstr-app-${app.id}`}
        onDeleted={handleDeleted}
      />
    </>
  );
}
