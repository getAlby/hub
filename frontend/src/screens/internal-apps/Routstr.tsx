import {
  ArrowLeftIcon,
  AlertTriangleIcon,
  CheckCircleIcon,
  ExternalLinkIcon,
} from "lucide-react";
import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { AboutAppCard } from "src/components/connections/AboutAppCard";
import { AppLinksCard } from "src/components/connections/AppLinksCard";
import { AppStoreDetailHeader } from "src/components/connections/AppStoreDetailHeader";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { InsufficientLightningBalanceAlert } from "src/components/InsufficientLightningBalanceAlert";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import ModelSelect from "src/components/connections/routstr/ModelSelect";
import { RoutstrConnectionDetails } from "src/components/connections/routstr/RoutstrConnectionDetails";
import { ROUTSTR_UNIVERSAL_KEY_COPY } from "src/components/connections/routstr/constants";
import PayFromSelect from "src/screens/wallet/send/PayFromSelect";
import { useBalances } from "src/hooks/useBalances";
import { useApps } from "src/hooks/useApps";
import {
  RoutstrdModel,
  createRoutstrdClient,
  getAllRoutstrdModels,
  getRoutstrdKeyBalances,
  fundFromHub,
} from "src/hooks/useRoutstrd";
import { useCapabilities } from "src/hooks/useCapabilities";
import {
  DEFAULT_APP_BUDGET_RENEWAL,
  DEFAULT_APP_BUDGET_SATS,
  PAY_FROM_SELECT_APPS_LIMIT,
} from "src/constants";
import { createApp } from "src/requests/createApp";
import { AppPermissions, Scope } from "src/types";
import Permissions from "src/components/Permissions";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

type WizardStep =
  | "start"
  | "configure"
  | "topup"
  | "create-key"
  | "fund-key"
  | "done";

const STEP_LABELS: Record<WizardStep, string> = {
  start: "",
  configure: "Configure",
  topup: "Top Up",
  "create-key": "Create Key",
  "fund-key": "Fund Key",
  done: "Done",
};

const STEP_ORDER: WizardStep[] = [
  "configure",
  "topup",
  "create-key",
  "fund-key",
  "done",
];

const ISOLATED_SCOPES: Scope[] = [
  "pay_invoice",
  "get_balance",
  "get_info",
  "make_invoice",
  "lookup_invoice",
  "list_transactions",
  "notifications",
];

function StepIndicator({ current }: { current: WizardStep }) {
  return (
    <div className="flex items-center gap-2 mb-6">
      {STEP_ORDER.map((s, i) => {
        const curIdx = STEP_ORDER.indexOf(current);
        const state = i < curIdx ? "done" : i === curIdx ? "active" : "pending";
        return (
          <React.Fragment key={s}>
            {i > 0 && (
              <div
                className={`flex-1 h-0.5 rounded ${
                  i <= curIdx ? "bg-primary" : "bg-muted"
                }`}
              />
            )}
            <div className="flex items-center gap-1.5">
              <div
                className={`flex items-center justify-center w-7 h-7 rounded-full text-xs font-bold ${
                  state === "done"
                    ? "bg-primary text-primary-foreground"
                    : state === "active"
                      ? "bg-primary text-primary-foreground ring-2 ring-primary/30"
                      : "bg-muted text-muted-foreground"
                }`}
              >
                {state === "done" ? (
                  <CheckCircleIcon className="w-4 h-4" />
                ) : (
                  i + 1
                )}
              </div>
              <span
                className={`text-sm hidden sm:inline ${
                  state === "pending"
                    ? "text-muted-foreground"
                    : "text-foreground"
                }`}
              >
                {STEP_LABELS[s]}
              </span>
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
}

export function Routstr() {
  const navigate = useNavigate();
  const [step, setStep] = useState<WizardStep>("start");
  const [appId, setAppId] = useState<number | null>(null);
  const [appPubkey, setAppPubkey] = useState<string | null>(null);

  // --- Step 1: Configure (uses standard Permissions component) ---
  const { data: capabilities } = useCapabilities();
  const [permissions, setPermissions] = useState<AppPermissions>({
    scopes: ISOLATED_SCOPES,
    maxAmountSat: DEFAULT_APP_BUDGET_SATS,
    budgetRenewal: DEFAULT_APP_BUDGET_RENEWAL,
    expiresAt: undefined,
    isolated: true,
  });
  const [configureLoading, setConfigureLoading] = useState(false);

  // --- Step 2: Top Up ---
  const { data: balances } = useBalances();
  const [topupAmount, setTopupAmount] = useState("5000");
  const [topupComment, setTopupComment] = useState("Routstr wallet top-up");
  const [payFromAppId, setPayFromAppId] = useState<number | undefined>(
    undefined
  );
  // Pay-from sources mirror PayFromSelect's list (main wallet + pay_invoice
  // apps). The balance line below the selector reflects the wallet the user
  // actually chose, and sufficiency is checked against that same wallet.
  const { data: payFromAppsData } = useApps(PAY_FROM_SELECT_APPS_LIMIT, 1, {
    name: "",
  });
  const selectedAppBalanceMsat = payFromAppsData?.apps?.find(
    (a) => a.id === payFromAppId
  )?.balanceMsat;
  const selectedBalanceMsat =
    payFromAppId === undefined
      ? balances?.lightning.totalSpendableMsat
      : selectedAppBalanceMsat;
  const [topupLoading, setTopupLoading] = useState(false);

  // --- Step 3: Create Key ---
  const [models, setModels] = useState<RoutstrdModel[]>([]);
  const [modelsLoading, setModelsLoading] = useState(false);
  const [keyLoading, setKeyLoading] = useState(false);

  // --- Step 4: Done ---
  const [createdApiKey, setCreatedApiKey] = useState("");
  const [createdClientId, setCreatedClientId] = useState("");
  const [fundedAmount, setFundedAmount] = useState(0);
  const [fundKeyAmount, setFundKeyAmount] = useState("5000");
  const [fundKeyLoading, setFundKeyLoading] = useState(false);

  const appStoreApp = appStoreApps.find((app) => app.id === "routstr");

  // ─── Handlers ───────────────────────────────────────────

  const handleConfigure = async (e: React.FormEvent) => {
    e.preventDefault();
    setConfigureLoading(true);
    try {
      const response = await createApp({
        name: appStoreApp?.title ?? "Routstr",
        scopes: permissions.scopes,
        maxAmountSat: permissions.maxAmountSat || 0,
        budgetRenewal: permissions.budgetRenewal,
        isolated: permissions.isolated,
        expiresAt: permissions.expiresAt?.toISOString(),
        metadata: { app_store_app_id: "routstr" },
      });
      setAppId(response.id);
      setAppPubkey(response.pairingPublicKey || null);
      toast.success("Routstr wallet created");

      // Register the NWC connection with routstrd (non-blocking)
      try {
        const { nwcConnect } = await import("src/hooks/useRoutstrd");
        await nwcConnect(response.pairingUri);
      } catch (e) {
        console.warn(
          "routstrd NWC connect failed (daemon may not be running):",
          e
        );
      }

      setStep("topup");
    } catch (error) {
      handleRequestError("Failed to create wallet", error);
    } finally {
      setConfigureLoading(false);
    }
  };

  const handleTopUp = async () => {
    const amount = Number(topupAmount);
    if (!amount || amount < 1) {
      toast.error("Enter a valid amount");
      return;
    }
    if (!appId) {
      toast.error("No wallet created yet. Go back to Configure.");
      return;
    }
    setTopupLoading(true);
    try {
      await request("/api/transfers", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          toAppId: appId,
          amountSat: amount,
          description: topupComment || "Routstr wallet top-up",
          // If an app wallet is selected as the source, draw from it;
          // otherwise the transfer comes from the main wallet.
          ...(payFromAppId !== undefined ? { fromAppId: payFromAppId } : {}),
        }),
      });
      toast.success(`Topped up ${amount} sats`);
      setStep("create-key");
    } catch (error) {
      handleRequestError("Top-up failed", error);
    } finally {
      setTopupLoading(false);
    }
  };

  const loadModels = useCallback(async (refresh = false) => {
    setModelsLoading(true);
    try {
      const result = await getAllRoutstrdModels(refresh);
      if (result?.models) {
        setModels(result.models);
      }
    } catch (e) {
      toast.error("Could not load models. Is routstrd running?");
      console.error(e);
    } finally {
      setModelsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (step === "create-key") {
      loadModels();
    }
  }, [step, loadModels]);

  const handleCreateKey = async () => {
    setKeyLoading(true);
    try {
      const clientName = `routstr-app-${appId || "wizard"}-${Date.now()}`;
      const result = await createRoutstrdClient(clientName);
      const apiKey = result?.client?.apiKey;
      const actualClientId = result?.client?.id || clientName;
      if (!apiKey) {
        throw new Error("Failed to create API key");
      }
      setCreatedApiKey(apiKey);
      setCreatedClientId(actualClientId);
      toast.success("API key created!");
      setStep("fund-key");
    } catch (error) {
      handleRequestError("Failed to create API key", error);
    } finally {
      setKeyLoading(false);
    }
  };

  const handleDone = async () => {
    // Save API key to app metadata before navigating
    if (appId && createdApiKey) {
      try {
        // Prefer pubkey from createApp; fall back to v2 apps API
        let pubkey = appPubkey;
        if (!pubkey) {
          const appResult = await request<{ appPubkey: string }>(
            `/api/v2/apps/${appId}`
          );
          pubkey = appResult?.appPubkey || null;
        }

        // Live wallet balance (key string holds no sats)
        let balance = fundedAmount;
        try {
          const bal = await getRoutstrdKeyBalances();
          if (typeof bal?.total === "number") {
            balance = bal.total;
          }
        } catch {
          /* keep fundedAmount */
        }

        if (pubkey) {
          // Merge with existing metadata so we don't wipe other fields
          let existingMeta: Record<string, unknown> = {
            app_store_app_id: "routstr",
          };
          try {
            const existing = await request<{
              metadata?: Record<string, unknown>;
            }>(`/api/v2/apps/${appId}`);
            if (existing?.metadata && typeof existing.metadata === "object") {
              existingMeta = { ...existing.metadata };
            }
          } catch {
            /* use defaults */
          }

          const routstrMeta: Record<string, unknown> = {
            apiKey: createdApiKey,
            clientId: createdClientId || `routstr-app-${appId}`,
            balance,
          };
          await request(`/api/apps/${pubkey}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              metadata: {
                ...existingMeta,
                app_store_app_id: "routstr",
                routstr: routstrMeta,
              },
            }),
          });
        } else {
          console.error("No app pubkey — cannot save Routstr metadata");
          toast.error("Could not save API key to connection metadata");
        }
      } catch (e) {
        console.error("Failed to save API key metadata", e);
        toast.error("Failed to save API key details");
      }
    }
    if (appId) {
      navigate(`/apps/${appId}`);
    } else {
      navigate("/apps?tab=app-store");
    }
  };

  const handleFundKey = async () => {
    const amount = Number(fundKeyAmount);
    if (!amount || amount < 1) {
      toast.error("Enter a valid deposit amount");
      return;
    }
    setFundKeyLoading(true);
    try {
      // Direct Hub-to-mint path — pay from this app's wallet
      await fundFromHub(amount, appId!);
      setFundedAmount(amount);
      toast.success(`Deposited ${amount} sats`);
      setFundKeyLoading(false);
      setStep("done");
    } catch (error) {
      handleRequestError("Deposit failed", error);
      setFundKeyLoading(false);
    }
  };

  // ─── Start (Zeus-style landing) ──────────────────────────  // ─── Start (Zeus-style landing) ──────────────────────────

  // All hooks above. Guard before any early return (React rules-of-hooks).
  if (!appStoreApp) {
    return null;
  }

  if (step === "start") {
    return (
      <div className="grid gap-3">
        <AppStoreDetailHeader
          appStoreApp={appStoreApp}
          contentRight={
            <Button onClick={() => setStep("configure")}>
              <NostrWalletConnectIcon className="w-4 h-4 mr-2" />
              Connect to Routstr
            </Button>
          }
        />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <AboutAppCard appStoreApp={appStoreApp} />
          <AppLinksCard appStoreApp={appStoreApp} />
        </div>
      </div>
    );
  }

  // ─── Wizard ─────────────────────────────────────────────

  return (
    <div className="grid gap-5 max-w-2xl">
      <StepIndicator current={step} />

      {step === "configure" && (
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">Configure Wallet</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleConfigure} className="flex flex-col gap-4">
              {capabilities ? (
                <Permissions
                  capabilities={capabilities}
                  permissions={permissions}
                  setPermissions={setPermissions}
                />
              ) : (
                <div className="py-4 text-sm text-muted-foreground">
                  Loading wallet capabilities...
                </div>
              )}

              <div className="flex gap-3 justify-end mt-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setStep("start")}
                >
                  <ArrowLeftIcon className="w-4 h-4 mr-2" />
                  Back
                </Button>
                <LoadingButton loading={configureLoading} type="submit">
                  Next
                </LoadingButton>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {step === "topup" && (
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">Top Up Wallet</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <p className="text-sm text-muted-foreground">
              Send sats to your Routstr wallet to fund API key creation.
            </p>

            {/* Pay From — 1st */}
            <PayFromSelect appId={payFromAppId} onChange={setPayFromAppId} />

            {/* Balance of the selected pay-from wallet — same presentation
                as the send-payment menu, but reflects the chosen source
                (main wallet vs app wallet) */}
            {selectedBalanceMsat !== undefined && (
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>
                  Lightning Balance:{" "}
                  <FormattedBitcoinAmount amountMsat={selectedBalanceMsat} />
                </span>
              </div>
            )}

            {/* Comment — 2nd */}
            <div className="grid gap-1.5">
              <Label htmlFor="topup-comment">Comment (optional)</Label>
              <Input
                id="topup-comment"
                value={topupComment}
                onChange={(e) => setTopupComment(e.target.value)}
                placeholder="What's this for?"
              />
            </div>

            {/* Amount — 3rd */}
            <div className="grid gap-1.5">
              <Label htmlFor="topup-amount">Amount (sats)</Label>
              <Input
                id="topup-amount"
                type="number"
                min={1}
                value={topupAmount}
                onChange={(e) => setTopupAmount(e.target.value)}
              />
              {Number(topupAmount) > 0 && (
                <div className="flex justify-end text-xs text-muted-foreground mt-1">
                  <FormattedFiatAmount amountSat={Number(topupAmount)} />
                </div>
              )}
            </div>

            {/* Sufficiency check against the SELECTED wallet: the shared
                InsufficientLightningBalanceAlert does main-wallet MPP math,
                so it only applies when the main wallet is the source; an app
                wallet is checked against its own balance */}
            {Number(topupAmount) > 0 &&
              balances &&
              payFromAppId === undefined && (
                <InsufficientLightningBalanceAlert
                  amountSat={Number(topupAmount)}
                />
              )}
            {Number(topupAmount) > 0 &&
              payFromAppId !== undefined &&
              selectedBalanceMsat !== undefined &&
              Number(topupAmount) * 1000 > selectedBalanceMsat && (
                <Alert>
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>Selected Wallet Balance Too Low</AlertTitle>
                  <AlertDescription>
                    The selected wallet has{" "}
                    <FormattedBitcoinAmount amountMsat={selectedBalanceMsat} />.
                    Send a smaller amount or pick another wallet.
                  </AlertDescription>
                </Alert>
              )}

            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setStep("configure")}>
                <ArrowLeftIcon className="w-4 h-4 mr-2" />
                Back
              </Button>
              <LoadingButton
                loading={topupLoading}
                onClick={handleTopUp}
                disabled={!topupAmount || Number(topupAmount) < 1}
              >
                Send {topupAmount} sats
              </LoadingButton>
              <Button variant="ghost" onClick={() => setStep("create-key")}>
                Next
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {step === "create-key" && (
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">Create API Key</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <>
              <p className="text-sm text-muted-foreground">
                {ROUTSTR_UNIVERSAL_KEY_COPY}
              </p>

              {/* Browse-only model dialog — clean trigger (no count, no
                  auto-routing subtitle); refreshes the list on open */}
              <ModelSelect
                models={models.filter((m) => m.enabled !== false)}
                browseOnly
                showCount={false}
                disabled={modelsLoading}
                onOpen={() => loadModels()}
              />

              <div className="flex gap-3 justify-end">
                <Button variant="outline" onClick={() => setStep("topup")}>
                  <ArrowLeftIcon className="w-4 h-4 mr-2" />
                  Back
                </Button>
                <LoadingButton loading={keyLoading} onClick={handleCreateKey}>
                  Create API Key
                </LoadingButton>
              </div>
            </>
          </CardContent>
        </Card>
      )}

      {step === "fund-key" && createdApiKey && (
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">Fund Key</CardTitle>
            <CardDescription>
              Your API key is free. Deposit sats to use it for AI inference.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <div>
              <Label>API Key</Label>
              <div className="mt-1 p-2 bg-muted rounded text-sm font-mono break-all select-all">
                {createdApiKey}
              </div>
            </div>

            <div className="grid gap-1.5">
              <Label htmlFor="fund-amount">Deposit (sats) — optional</Label>
              <Input
                id="fund-amount"
                type="number"
                min={1}
                value={fundKeyAmount}
                onChange={(e) => setFundKeyAmount(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                You can also fund your key later on the connection page.
              </p>
            </div>

            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setStep("done")}>
                Skip
              </Button>
              <LoadingButton
                loading={fundKeyLoading}
                onClick={handleFundKey}
                disabled={!fundKeyAmount || Number(fundKeyAmount) < 1}
              >
                Deposit & Finish
              </LoadingButton>
            </div>
          </CardContent>
        </Card>
      )}

      {step === "done" && (
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl flex items-center gap-2">
              <CheckCircleIcon className="text-green-500" />
              All Set!
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <p className="text-muted-foreground">
              Your Routstr connection is ready. Use these details with any
              OpenAI-compatible client.
            </p>

            {createdApiKey && (
              <RoutstrConnectionDetails apiKey={createdApiKey} showHubUrl />
            )}

            <Button onClick={handleDone} className="w-fit">
              <ExternalLinkIcon className="mr-2 h-4 w-4" />
              Open Connection Detail
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
