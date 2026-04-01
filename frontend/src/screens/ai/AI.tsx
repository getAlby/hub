import {
  ArrowRightIcon,
  ArrowUpRightIcon,
  BotIcon,
  BoxIcon,
  CheckCircleIcon,
  CopyIcon,
  HammerIcon,
  InfoIcon,
  LayersIcon,
  type LucideIcon,
  RepeatIcon,
  ShoppingBagIcon,
  SparklesIcon,
  XIcon,
  ZapIcon,
} from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import bitrefillLogo from "src/assets/suggested-apps/bitrefill.png";
import claudeLogo from "src/assets/suggested-apps/claude.png";
import clineLogo from "src/assets/suggested-apps/cline.png";
import codexLogo from "src/assets/suggested-apps/codex.png";
import cursorLogo from "src/assets/suggested-apps/cursor.png";
import gooseLogo from "src/assets/suggested-apps/goose.png";
import openclawLogo from "src/assets/suggested-apps/openclaw.png";
import opencodeLogo from "src/assets/suggested-apps/opencode.png";
import payperqLogo from "src/assets/suggested-apps/payperq.png";
import AppHeader from "src/components/AppHeader";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "src/components/ui/tabs";
import {
  DEFAULT_APP_BUDGET_RENEWAL,
  DEFAULT_APP_BUDGET_SATS,
  localStorageKeys,
} from "src/constants";
import { useApp } from "src/hooks/useApp";
import { useAppsForAppStoreApp } from "src/hooks/useApps";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

type Agent = {
  id: string;
  name: string;
  logo: string;
  description: string;
  setupUrl: string;
};

const agents: Agent[] = [
  {
    id: "openclaw",
    name: "OpenClaw",
    logo: openclawLogo,
    description: "Open-source AI coding agent",
    setupUrl: "",
  },
  {
    id: "cursor",
    name: "Cursor",
    logo: cursorLogo,
    description: "AI-powered code editor",
    setupUrl: "",
  },
  {
    id: "claude",
    name: "Claude",
    logo: claudeLogo,
    description: "AI assistant for conversations, analysis, and coding",
    setupUrl: "/internal-apps/claude",
  },
  {
    id: "codex",
    name: "Codex",
    logo: codexLogo,
    description: "OpenAI's coding agent",
    setupUrl: "",
  },
  {
    id: "cline",
    name: "Cline",
    logo: clineLogo,
    description: "AI coding assistant for VS Code",
    setupUrl: "",
  },
  {
    id: "goose",
    name: "Goose",
    logo: gooseLogo,
    description: "Local AI agent by Block for automating engineering tasks",
    setupUrl: "/internal-apps/goose",
  },
  {
    id: "opencode",
    name: "OpenCode",
    logo: opencodeLogo,
    description: "Terminal-based AI coding assistant",
    setupUrl: "",
  },
];

export function AI() {
  const navigate = useNavigate();
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const [createdAppId, setCreatedAppId] = React.useState<number>();
  const [expandedAgent, setExpandedAgent] = React.useState<string | null>(null);
  const [heroDismissed, setHeroDismissed] = React.useState(
    () => localStorage.getItem(localStorageKeys.aiHeroDismissed) === "true"
  );

  // Poll the created app to detect when the agent connects
  const { data: createdApp } = useApp(createdAppId, !!createdAppId);
  const isConnected = !!createdApp?.lastUsedAt;

  const claudeApp = appStoreApps.find((app) => app.id === "claude");
  const gooseApp = appStoreApps.find((app) => app.id === "goose");

  const claudeConnections = useAppsForAppStoreApp(claudeApp!);
  const gooseConnections = useAppsForAppStoreApp(gooseApp!);

  const dismissHero = React.useCallback(() => {
    setHeroDismissed(true);
    localStorage.setItem(localStorageKeys.aiHeroDismissed, "true");
  }, []);

  const handleCreateConnection = async (agentId: string) => {
    setLoading(true);
    try {
      const agent = agents.find((a) => a.id === agentId);
      const agentName = agent?.name ?? "AI Agent";
      const createAppResponse = await createApp({
        name: agentName,
        scopes: [
          "get_info",
          "get_balance",
          "list_transactions",
          "lookup_invoice",
          "make_invoice",
          "notifications",
          "pay_invoice",
          "sign_message",
        ],
        maxAmount: DEFAULT_APP_BUDGET_SATS,
        budgetRenewal: DEFAULT_APP_BUDGET_RENEWAL,
        metadata: {
          app_store_app_id: agentId,
        },
      });
      setConnectionSecret(createAppResponse.pairingUri);
      setCreatedAppId(createAppResponse.id);
      setExpandedAgent(agentId);
      toast(`${agentName} connection created`);
    } catch (error) {
      handleRequestError("Failed to create connection", error);
    }
    setLoading(false);
  };

  const mcpUrl = `https://mcp.getalby.com/mcp`;
  const mcpUrlWithSecret = `${mcpUrl}?nwc=${encodeURIComponent(connectionSecret)}`;
  const gooseDesktopLink = `goose://extension?transport=streamable_http&url=${encodeURIComponent(mcpUrlWithSecret)}&id=alby&name=Alby&description=Connect%20Goose%20to%20your%20Bitcoin%20Lightning%20Wallet`;

  return (
    <>
      <AppHeader title="AI & Agents" pageTitle="AI & Agents" />

      {/* Hero — collapsible, persisted in localStorage */}
      {!heroDismissed && (
        <div className="overflow-hidden">
          <div className="bg-card text-card-foreground rounded-xl overflow-hidden relative border border-border">
            <button
              onClick={dismissHero}
              className="absolute top-4 right-4 z-10 text-muted-foreground hover:text-foreground transition-colors"
              aria-label="Dismiss"
            >
              <XIcon className="w-5 h-5" />
            </button>
            <div className="flex flex-col lg:flex-row min-h-[360px]">
              {/* Left */}
              <div className="flex-1 p-8 lg:p-12 flex flex-col justify-center">
                <p className="text-muted-foreground text-sm font-medium tracking-widest uppercase mb-4">
                  Agentic Payments
                </p>
                <h2 className="text-4xl lg:text-5xl font-bold tracking-tight leading-[1.1] mb-4">
                  Your AI agent,
                  <br />
                  <span className="text-primary">powered by bitcoin</span>
                </h2>
                <p className="text-muted-foreground text-lg max-w-md">
                  Connect any AI agent and let it pay for services, buy gift
                  cards, trade, and access 16,000+ paid APIs.
                </p>
              </div>

              {/* Right — terminal mockup */}
              <div className="flex-1 p-6 pt-12 lg:p-8 lg:pt-14 flex items-center">
                <div className="w-full max-w-2xl ml-auto bg-muted rounded-lg border border-border overflow-hidden shadow-2xl">
                  <div className="flex items-center gap-1.5 px-4 py-2.5 border-b border-border">
                    <div className="w-2.5 h-2.5 rounded-full bg-muted-foreground/30" />
                    <div className="w-2.5 h-2.5 rounded-full bg-muted-foreground/30" />
                    <div className="w-2.5 h-2.5 rounded-full bg-muted-foreground/30" />
                    <span className="text-muted-foreground text-xs ml-2 font-mono">
                      claude
                    </span>
                  </div>
                  <div className="p-4 font-mono text-sm space-y-3">
                    <div>
                      <span className="text-muted-foreground">&gt; </span>
                      <span className="text-foreground">
                        Buy a $15 DoorDash gift card
                      </span>
                    </div>
                    <div className="text-muted-foreground pl-3 border-l-2 border-primary/40 space-y-1">
                      <p>
                        <SparklesIcon className="w-3 h-3 inline text-primary mr-1" />
                        Searching Bitrefill for DoorDash...
                      </p>
                      <p>
                        <ZapIcon className="w-3 h-3 inline text-primary mr-1" />
                        Paying 45,210 sats via Lightning
                      </p>
                      <p>
                        <CheckCircleIcon className="w-3 h-3 inline text-positive-foreground mr-1" />
                        <span className="text-positive-foreground">Done!</span>{" "}
                        Gift card code: XXXX-XXXX-XXXX
                      </p>
                    </div>
                    <div>
                      <span className="text-muted-foreground">&gt; </span>
                      <span className="text-foreground">
                        Send $5 to hub@getalby.com
                      </span>
                    </div>
                    <div className="text-muted-foreground pl-3 border-l-2 border-primary/40 space-y-1">
                      <p>
                        <ZapIcon className="w-3 h-3 inline text-primary mr-1" />
                        Sending 15,000 sats...
                      </p>
                      <p>
                        <CheckCircleIcon className="w-3 h-3 inline text-positive-foreground mr-1" />
                        <span className="text-positive-foreground">Sent!</span>{" "}
                        Payment confirmed
                      </p>
                    </div>
                    <div className="flex items-center">
                      <span className="text-muted-foreground">&gt; </span>
                      <span className="w-2 h-4 bg-primary ml-0.5 animate-[blink_1s_step-end_infinite]" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* How it works — 3 steps */}
            <div className="border-t border-border grid grid-cols-1 sm:grid-cols-3">
              <div className="p-6 lg:p-8 border-b sm:border-b-0 sm:border-r border-border">
                <div className="text-primary font-mono text-sm mb-2">01</div>
                <h3 className="font-semibold text-lg mb-1">Connect</h3>
                <p className="text-muted-foreground text-sm">
                  Create a connection and paste it into OpenClaw, Cursor,
                  Claude, or any AI agent.
                </p>
              </div>
              <div className="p-6 lg:p-8 border-b sm:border-b-0 sm:border-r border-border">
                <div className="text-primary font-mono text-sm mb-2">02</div>
                <h3 className="font-semibold text-lg mb-1">Prompt</h3>
                <p className="text-muted-foreground text-sm">
                  Ask your agent to buy, pay, trade, or access any paid API
                  using natural language.
                </p>
              </div>
              <div className="p-6 lg:p-8">
                <div className="text-primary font-mono text-sm mb-2">03</div>
                <h3 className="font-semibold text-lg mb-1">Pay</h3>
                <p className="text-muted-foreground text-sm">
                  Your agent pays instantly via Lightning. Set budgets and
                  permissions per agent.
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Connect section */}
      <div className="space-y-4">
        <Card className="border-primary bg-primary/10">
          <CardHeader>
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-primary/20 flex items-center justify-center shrink-0">
                  <BotIcon className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <CardTitle>Connect Your Agent</CardTitle>
                  <p className="text-sm text-muted-foreground mt-1">
                    Create a connection and paste it into your AI agent to get
                    started
                  </p>
                </div>
              </div>
              <Link
                to="/apps?tab=app-store&category=ai"
                className="text-xs text-muted-foreground hover:text-foreground transition-colors flex items-center gap-1"
              >
                All AI Apps
                <ArrowRightIcon className="w-3 h-3" />
              </Link>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-2">
              <Select
                value={expandedAgent ?? undefined}
                onValueChange={(value) => {
                  setConnectionSecret("");
                  setCreatedAppId(undefined);
                  setExpandedAgent(value);
                }}
              >
                <SelectTrigger className="w-[240px]">
                  <SelectValue
                    placeholder={
                      <span className="flex items-center gap-2">
                        <span className="flex -space-x-2">
                          {agents.slice(0, 4).map((agent) => (
                            <img
                              key={agent.id}
                              src={agent.logo}
                              alt={agent.name}
                              className="w-5 h-5 rounded-full border-2 border-background"
                            />
                          ))}
                        </span>
                        <span className="text-muted-foreground">
                          Choose your agent
                        </span>
                      </span>
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {agents.map((agent) => (
                    <SelectItem key={agent.id} value={agent.id}>
                      <div className="flex items-center gap-2">
                        <img
                          src={agent.logo}
                          alt={agent.name}
                          className="w-5 h-5 rounded"
                        />
                        {agent.name}
                      </div>
                    </SelectItem>
                  ))}
                  <SelectItem value="other">Other</SelectItem>
                </SelectContent>
              </Select>
              <Button
                onClick={() => {
                  if (!expandedAgent) {
                    return;
                  }
                  const agent = agents.find((a) => a.id === expandedAgent);
                  const connections =
                    expandedAgent === "claude"
                      ? claudeConnections
                      : expandedAgent === "goose"
                        ? gooseConnections
                        : undefined;
                  const hasExistingConnection = (connections?.length ?? 0) > 0;
                  if (hasExistingConnection && agent?.setupUrl) {
                    navigate(agent.setupUrl);
                  } else {
                    handleCreateConnection(expandedAgent);
                  }
                }}
                disabled={isLoading || !expandedAgent}
              >
                <ZapIcon className="w-4 h-4" />
                Connect
              </Button>
            </div>

            {/* Connection state */}
            {expandedAgent &&
              connectionSecret &&
              (isConnected ? (
                <div className="flex items-center gap-2 text-sm">
                  <CheckCircleIcon className="w-4 h-4 text-positive-foreground" />
                  <span className="text-positive-foreground font-medium">
                    Agent connected
                  </span>
                  <Link
                    to={`/apps/${createdAppId}`}
                    className="text-muted-foreground hover:text-foreground underline transition-colors"
                  >
                    Set name & manage
                  </Link>
                </div>
              ) : (
                <>
                  <ConnectionInstructions
                    agentId={expandedAgent}
                    connectionSecret={connectionSecret}
                    mcpUrl={mcpUrl}
                    mcpUrlWithSecret={mcpUrlWithSecret}
                    gooseDesktopLink={gooseDesktopLink}
                  />
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Loading className="size-4" />
                    Waiting for agent to connect...
                  </div>
                </>
              ))}
          </CardContent>
        </Card>

        {/* Inspiration */}
        <InspirationPrompts />

        {/* Featured services — branded full cards */}
        <div>
          <h2 className="text-2xl font-bold mb-4">Featured Services</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            <ExternalLink to="https://www.bitrefill.com/agents">
              <Card className="group relative h-full hover:border-primary/30 transition-colors p-0">
                <CardContent className="p-4 flex flex-col h-full">
                  <div className="flex items-center gap-3 mb-3">
                    <img
                      src={bitrefillLogo}
                      alt="Bitrefill"
                      className="w-10 h-10 rounded-lg"
                    />
                    <div className="flex-1">
                      <p className="font-semibold">Bitrefill</p>
                      <p className="text-xs text-muted-foreground">
                        Gift cards & e-SIMs
                      </p>
                    </div>
                  </div>
                  <ArrowUpRightIcon className="w-4 h-4 text-muted-foreground/0 group-hover:text-muted-foreground transition-colors absolute top-4 right-4" />
                  <p className="text-sm text-muted-foreground flex-1">
                    &quot;Buy a $15 DoorDash gift card&quot; — your agent pays
                    via Lightning and delivers the code.
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-3 font-mono">
                    MCP Server
                  </p>
                </CardContent>
              </Card>
            </ExternalLink>

            <ExternalLink to="https://ppq.ai/invite/3f21c1e5">
              <Card className="group relative h-full hover:border-primary/30 transition-colors p-0">
                <CardContent className="p-4 flex flex-col h-full">
                  <div className="flex items-center gap-3 mb-3">
                    <img
                      src={payperqLogo}
                      alt="PPQ.ai"
                      className="w-10 h-10 rounded-lg"
                    />
                    <div className="flex-1">
                      <p className="font-semibold">PPQ.ai</p>
                      <p className="text-xs text-muted-foreground">
                        AI model access
                      </p>
                    </div>
                  </div>
                  <ArrowUpRightIcon className="w-4 h-4 text-muted-foreground/0 group-hover:text-muted-foreground transition-colors absolute top-4 right-4" />
                  <p className="text-sm text-muted-foreground flex-1">
                    Pay-per-prompt access to top AI models. No subscription —
                    pay with Lightning.
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-3 font-mono">
                    NWC Auto Top-up
                  </p>
                </CardContent>
              </Card>
            </ExternalLink>

            <ExternalLink to="https://402index.io">
              <Card className="group relative h-full hover:border-primary/30 transition-colors p-0">
                <CardContent className="p-4 flex flex-col h-full">
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-10 h-10 rounded-lg bg-[#7c8aff]/10 flex items-center justify-center">
                      <LayersIcon className="w-5 h-5 text-[#7c8aff]" />
                    </div>
                    <div className="flex-1">
                      <p className="font-semibold">402 Index</p>
                      <p className="text-xs text-muted-foreground">
                        16,000+ APIs
                      </p>
                    </div>
                  </div>
                  <ArrowUpRightIcon className="w-4 h-4 text-muted-foreground/0 group-hover:text-muted-foreground transition-colors absolute top-4 right-4" />
                  <p className="text-sm text-muted-foreground flex-1">
                    Directory of paid API endpoints — search, data, compute,
                    LLMs, and more.
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-3 font-mono">
                    L402 / x402 / MPP
                  </p>
                </CardContent>
              </Card>
            </ExternalLink>
          </div>
        </div>
      </div>
    </>
  );
}

function ConnectionInstructions({
  agentId,
  connectionSecret,
  mcpUrl,
  mcpUrlWithSecret,
  gooseDesktopLink,
}: {
  agentId: string;
  connectionSecret: string;
  mcpUrl: string;
  mcpUrlWithSecret: string;
  gooseDesktopLink: string;
}) {
  if (agentId === "claude") {
    return (
      <Accordion type="single" collapsible>
        <AccordionItem value="web">
          <AccordionTrigger>Claude Web</AccordionTrigger>
          <AccordionContent>
            <ol className="list-decimal list-inside space-y-1 text-sm">
              <li>
                Visit{" "}
                <a
                  href="https://claude.ai"
                  target="_blank"
                  className="underline"
                >
                  claude.ai
                </a>{" "}
                and sign in
              </li>
              <li>Go to Settings &rarr; Integrations</li>
              <li>Click +Add integration</li>
              <li>Integration Name: Alby</li>
              <li>
                Paste the integration URL:{" "}
                <Button
                  onClick={() => copyToClipboard(mcpUrlWithSecret)}
                  size="sm"
                  variant="secondary"
                >
                  Copy URL
                </Button>
              </li>
            </ol>
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="desktop">
          <AccordionTrigger>Claude Desktop</AccordionTrigger>
          <AccordionContent>
            <ol className="list-decimal list-inside space-y-1 text-sm">
              <li>
                Download{" "}
                <a
                  href="https://claude.ai/download"
                  target="_blank"
                  className="underline"
                >
                  Claude Desktop
                </a>
              </li>
              <li>Open Claude Desktop and sign in</li>
              <li>Go to Settings &rarr; Integrations</li>
              <li>Click +Add integration</li>
              <li>Integration Name: Alby</li>
              <li>
                Paste the integration URL:{" "}
                <Button
                  onClick={() => copyToClipboard(mcpUrlWithSecret)}
                  size="sm"
                  variant="secondary"
                >
                  Copy URL
                </Button>
              </li>
            </ol>
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="code">
          <AccordionTrigger>Claude Code</AccordionTrigger>
          <AccordionContent>
            <ol className="list-decimal list-inside space-y-1 text-sm">
              <li>
                Install{" "}
                <a
                  href="https://www.anthropic.com/claude-code"
                  target="_blank"
                  className="underline"
                >
                  Claude Code
                </a>
              </li>
              <li>
                Run this command in your terminal:{" "}
                <Button
                  onClick={() =>
                    copyToClipboard(
                      `claude mcp add --transport http alby https://mcp.getalby.com/mcp --header "Authorization: Bearer ${connectionSecret}"`
                    )
                  }
                  size="sm"
                  variant="secondary"
                >
                  Copy command
                </Button>
              </li>
            </ol>
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    );
  }

  if (agentId === "goose") {
    return (
      <Accordion type="single" collapsible>
        <AccordionItem value="desktop">
          <AccordionTrigger>Goose Desktop</AccordionTrigger>
          <AccordionContent>
            <a href={gooseDesktopLink}>
              <Button>Connect to Goose Desktop</Button>
            </a>
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="cli">
          <AccordionTrigger>Goose CLI</AccordionTrigger>
          <AccordionContent>
            <ol className="list-decimal list-inside space-y-1 text-sm">
              <li>
                Run <span className="font-semibold">goose configure</span>
              </li>
              <li>
                Choose{" "}
                <span className="font-semibold">
                  Remote Extension (Streaming HTTP)
                </span>
              </li>
              <li>
                Name: <span className="font-semibold">Alby</span>
              </li>
              <li>
                Endpoint URI:{" "}
                <Button
                  onClick={() => copyToClipboard(mcpUrl)}
                  size="sm"
                  variant="secondary"
                >
                  Copy URI
                </Button>
              </li>
              <li>
                Timeout: <span className="font-semibold">300</span>
              </li>
              <li>
                Add custom headers: <span className="font-semibold">yes</span>
              </li>
              <li>
                Header name:{" "}
                <span className="font-semibold">Authorization</span>
              </li>
              <li>
                Header value:{" "}
                <Button
                  onClick={() => copyToClipboard(`Bearer ${connectionSecret}`)}
                  size="sm"
                  variant="secondary"
                >
                  Copy value
                </Button>
              </li>
            </ol>
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    );
  }

  // Generic fallback — a human-readable prompt the user copies into their agent
  const promptTemplate = (secret: string) =>
    `Install the skill from https://getalby.com/cli/SKILL.md and use the setup command with this connection secret: ${secret}`;
  const genericPrompt = promptTemplate(connectionSecret);
  const redactedPrompt = promptTemplate(
    `${connectionSecret.slice(0, 30)}${"•".repeat(20)}`
  );

  return (
    <div className="space-y-3 text-sm">
      <p className="text-muted-foreground">
        Copy this prompt and paste it into your agent:
      </p>
      <button
        onClick={() => copyToClipboard(genericPrompt)}
        className="flex items-center gap-3 rounded-lg bg-muted/50 border border-border px-4 py-3 w-full text-left cursor-pointer hover:border-primary/30 transition-colors"
      >
        <ArrowRightIcon className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
        <p className="flex-1 text-sm font-mono break-all select-none">
          {redactedPrompt}
        </p>
        <CopyIcon className="w-4 h-4 text-muted-foreground shrink-0" />
      </button>
    </div>
  );
}

const inspirationCategories: {
  label: string;
  icon: LucideIcon;
  prompts: string[];
  skill?: { prompt: string; skillName: string; url: string };
}[] = [
  {
    label: "Payments",
    icon: ZapIcon,
    prompts: [
      "send $5 to hub@getalby.com for coffee",
      "pay this lightning invoice: lnbc1pjk...",
      "send 1,000 sats to each of these 20 podcast hosts",
      "split last night's dinner bill equally between these 4 lightning addresses",
    ],
  },
  {
    label: "Shopping",
    icon: ShoppingBagIcon,
    prompts: [
      "buy a $25 Netflix gift card",
      "get me an eSIM with 5GB of data for my trip to Portugal",
      "find me a VPN and a phone number, no KYC",
    ],
  },
  {
    label: "Automation",
    icon: RepeatIcon,
    prompts: [
      "read invoices.csv and pay all 47 lightning invoices",
      "check my balance every morning and message me if it drops below 100k sats",
      "every Monday, send 50,000 sats to hub@getalby.com",
    ],
  },
  {
    label: "Build Bitcoin Apps",
    icon: HammerIcon,
    prompts: [
      "build a paywall for my API that charges 10 sats per request",
      "scaffold a Next.js app with lightning login using Alby JS SDK",
      "create a tipping page that generates invoices on the fly",
      "add a lightning paywall to my Express API using L402",
    ],
    skill: {
      prompt: "Install the skill from https://github.com/getAlby/builder-skill",
      skillName: "Builder Skill",
      url: "https://github.com/getAlby/builder-skill",
    },
  },
  {
    label: "Node Management",
    icon: BoxIcon,
    prompts: [
      "open a channel with 2M sats to ACINQ's node",
      "show me my top 5 channels by routing volume this month",
      "close all channels with less than 10k sats capacity",
      "what's my on-chain vs lightning balance breakdown?",
    ],
    skill: {
      prompt: "Install the skill from https://github.com/getAlby/hub-skill",
      skillName: "Alby Hub Skill",
      url: "https://github.com/getAlby/hub-skill",
    },
  },
];

function RotatingPrompt({ prompts }: { prompts: string[] }) {
  const [index, setIndex] = React.useState(0);
  const [charCount, setCharCount] = React.useState(0);
  const [isTyping, setIsTyping] = React.useState(true);

  const currentPrompt = prompts[index];

  // Reset when prompts change (tab switch)
  React.useEffect(() => {
    setIndex(0);
    setCharCount(0);
    setIsTyping(true);
  }, [prompts]);

  // Typing effect — advance characters
  React.useEffect(() => {
    if (!isTyping) {
      return;
    }
    if (charCount >= currentPrompt.length) {
      setIsTyping(false);
      return;
    }
    const timeout = setTimeout(() => {
      setCharCount((c) => c + 1);
    }, 30);
    return () => clearTimeout(timeout);
  }, [charCount, currentPrompt, isTyping]);

  // Pause then rotate to next prompt
  React.useEffect(() => {
    if (isTyping) {
      return;
    }
    const pause = setTimeout(() => {
      setIndex((i) => (i + 1) % prompts.length);
      setCharCount(0);
      setIsTyping(true);
    }, 3000);
    return () => clearTimeout(pause);
  }, [isTyping, prompts]);

  return (
    <div className="flex items-center gap-3 rounded-lg bg-muted/50 border border-border px-4 py-3">
      <span className="text-muted-foreground select-none">&rsaquo;</span>
      <p className="flex-1 text-sm font-mono">
        {currentPrompt.slice(0, charCount)}
        <span
          className={`inline-block w-2 h-[1.1em] translate-y-[2px] ml-0.5 bg-primary ${
            isTyping ? "" : "animate-[blink_1s_step-end_infinite]"
          }`}
        />
      </p>
      <button
        onClick={() => copyToClipboard(currentPrompt)}
        className="text-muted-foreground hover:text-foreground transition-colors shrink-0"
        aria-label="Copy prompt"
      >
        <CopyIcon className="w-4 h-4" />
      </button>
    </div>
  );
}

function InspirationPrompts() {
  return (
    <Tabs defaultValue={inspirationCategories[0].label} variant="line">
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-5 pt-5 pb-4">
          <p className="font-semibold text-sm mb-4">
            Give your agent new capabilities
          </p>
          <TabsList>
            {inspirationCategories.map((cat) => {
              const Icon = cat.icon;
              return (
                <TabsTrigger
                  key={cat.label}
                  value={cat.label}
                  className="gap-1.5"
                >
                  <Icon className="w-3.5 h-3.5 translate-y-px" />
                  {cat.label}
                </TabsTrigger>
              );
            })}
          </TabsList>
        </div>

        <div className="px-5 pb-5">
          {inspirationCategories.map((cat) => (
            <TabsContent
              key={cat.label}
              value={cat.label}
              className="mt-0 space-y-2"
            >
              <RotatingPrompt prompts={cat.prompts} />
              {cat.skill && (
                <div className="flex items-center gap-2 px-1 text-xs text-muted-foreground">
                  <InfoIcon className="w-3.5 h-3.5 shrink-0" />
                  <span className="flex-1">
                    Requires the{" "}
                    <ExternalLink
                      to={cat.skill.url}
                      className="underline font-medium text-foreground hover:text-primary transition-colors"
                    >
                      {cat.skill.skillName}
                    </ExternalLink>{" "}
                    &mdash; ask your agent:{" "}
                    <em>&ldquo;{cat.skill.prompt}&rdquo;</em>
                  </span>
                </div>
              )}
            </TabsContent>
          ))}
        </div>
      </div>
    </Tabs>
  );
}
