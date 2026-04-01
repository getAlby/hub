import {
  ArrowRightIcon,
  ArrowUpRightIcon,
  CheckCircleIcon,
  CopyIcon,
  LayersIcon,
  PlusIcon,
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
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "src/components/ui/tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import {
  DEFAULT_APP_BUDGET_RENEWAL,
  DEFAULT_APP_BUDGET_SATS,
  localStorageKeys,
} from "src/constants";
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
    id: "claude",
    name: "Claude",
    logo: claudeLogo,
    description: "AI assistant for conversations, analysis, and coding",
    setupUrl: "/internal-apps/claude",
  },
  {
    id: "goose",
    name: "Goose",
    logo: gooseLogo,
    description: "Local AI agent by Block for automating engineering tasks",
    setupUrl: "/internal-apps/goose",
  },
];

export function AI() {
  const navigate = useNavigate();
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const [expandedAgent, setExpandedAgent] = React.useState<string | null>(null);
  const [heroDismissed, setHeroDismissed] = React.useState(
    () => localStorage.getItem(localStorageKeys.aiHeroDismissed) === "true"
  );

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
      const agentName =
        agentId === "claude"
          ? "Claude"
          : agentId === "goose"
            ? "Goose"
            : "AI Agent";
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
                        Send $5 to rene@getalby.com
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
                      <span className="w-2 h-4 bg-primary animate-pulse ml-0.5" />
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
                  Create a connection and paste it into Claude, Goose, or any
                  MCP agent.
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
        {!heroDismissed && (
          <h2 className="text-2xl font-bold mt-4">Connect Your Agent</h2>
        )}

        {/* Agent grid — flat, no nesting */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          {agents.map((agent) => {
            const connections =
              agent.id === "claude" ? claudeConnections : gooseConnections;
            const isConnected = (connections?.length ?? 0) > 0;

            return (
              <button
                key={agent.id}
                onClick={() =>
                  isConnected
                    ? navigate(agent.setupUrl)
                    : handleCreateConnection(agent.id)
                }
                disabled={isLoading}
                className="group relative rounded-xl border bg-card p-4 text-left hover:border-primary/40 transition-colors disabled:opacity-50"
              >
                {isConnected && (
                  <Badge
                    variant="positive"
                    className="absolute top-3 right-3 flex items-center gap-1 text-[10px]"
                  >
                    <CheckCircleIcon className="w-2.5 h-2.5" />
                    Connected
                  </Badge>
                )}
                <img
                  src={agent.logo}
                  alt={agent.name}
                  className="w-10 h-10 rounded-lg mb-3"
                />
                <p className="font-semibold text-sm">{agent.name}</p>
                <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                  {agent.description}
                </p>
              </button>
            );
          })}

          {/* Generic: Any MCP Agent */}
          <button
            onClick={() =>
              connectionSecret && expandedAgent === "other"
                ? undefined
                : handleCreateConnection("other")
            }
            disabled={isLoading}
            className="group rounded-xl border border-dashed bg-card p-4 text-left hover:border-primary/40 transition-colors disabled:opacity-50"
          >
            <div className="flex items-center mb-3">
              <Avatar className="size-10 rounded-lg border-2 border-card">
                <AvatarImage src={clineLogo} alt="Cline" />
              </Avatar>
              <Avatar className="size-10 rounded-lg border-2 border-card -ml-3">
                <AvatarImage src={codexLogo} alt="Codex" />
              </Avatar>
              <Avatar className="size-10 rounded-lg border-2 border-card -ml-3">
                <AvatarImage src={cursorLogo} alt="Cursor" />
              </Avatar>
              <Avatar className="size-10 rounded-lg border-2 border-card -ml-3">
                <AvatarImage src={openclawLogo} alt="OpenClaw" />
              </Avatar>
              <Avatar className="size-10 rounded-lg border-2 border-card -ml-3">
                <AvatarImage src={opencodeLogo} alt="OpenCode" />
              </Avatar>
              <Avatar className="size-10 rounded-lg border-2 border-card -ml-3 bg-muted">
                <AvatarFallback className="rounded-lg">
                  <PlusIcon className="w-5 h-5 text-muted-foreground" />
                </AvatarFallback>
              </Avatar>
            </div>
            <p className="font-semibold text-sm">Other Agent</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Cline, Codex, Cursor, OpenClaw, OpenCode & others
            </p>
          </button>

          {/* Browse all AI apps */}
          <Link
            to="/apps?tab=app-store&category=ai"
            className="group rounded-xl border border-dashed bg-card p-4 text-left hover:border-primary/40 transition-colors"
          >
            <div className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center mb-3">
              <SparklesIcon className="w-5 h-5 text-muted-foreground" />
            </div>
            <p className="font-semibold text-sm">All AI Apps</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Browse the full AI app store
            </p>
          </Link>
        </div>

        {/* Connection setup — shown after clicking an agent */}
        {expandedAgent && connectionSecret && (
          <Card>
            <CardHeader>
              <div className="flex items-center gap-3">
                {expandedAgent !== "other" && (
                  <img
                    src={expandedAgent === "claude" ? claudeLogo : gooseLogo}
                    className="w-10 h-10 rounded-lg"
                  />
                )}
                <div>
                  <CardTitle>
                    Connect{" "}
                    {expandedAgent === "claude"
                      ? "Claude"
                      : expandedAgent === "goose"
                        ? "Goose"
                        : "Your Agent"}
                  </CardTitle>
                  <CardDescription>
                    Follow the steps below to complete the connection.
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <ConnectionInstructions
                agentId={expandedAgent}
                connectionSecret={connectionSecret}
                mcpUrl={mcpUrl}
                mcpUrlWithSecret={mcpUrlWithSecret}
                gooseDesktopLink={gooseDesktopLink}
              />
            </CardContent>
          </Card>
        )}

        {/* Inspiration */}
        <InspirationPrompts />

        {/* Featured services — branded full cards */}
        <div>
          <h2 className="text-2xl font-bold mb-4">Featured Services</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            <ExternalLink to="https://www.bitrefill.com/account/developers/mcp-server">
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

            <ExternalLink to="https://nadanada.me">
              <Card className="group relative h-full hover:border-primary/30 transition-colors p-0">
                <CardContent className="p-4 flex flex-col h-full">
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-10 h-10 rounded-lg bg-[#ffd700] flex items-center justify-center">
                      <span className="text-black font-black text-[10px] leading-none text-center">
                        nada
                        <br />
                        nada
                      </span>
                    </div>
                    <div className="flex-1">
                      <p className="font-semibold">nadanada</p>
                      <p className="text-xs text-muted-foreground">
                        VPN, eSIM & phone numbers
                      </p>
                    </div>
                  </div>
                  <ArrowUpRightIcon className="w-4 h-4 text-muted-foreground/0 group-hover:text-muted-foreground transition-colors absolute top-4 right-4" />
                  <p className="text-sm text-muted-foreground flex-1">
                    Anonymous VPN, eSIM, and phone numbers. No account, no KYC —
                    pay with Lightning.
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-3 font-mono">
                    Lightning Payments
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
  if (agentId === "other") {
    return (
      <Tabs defaultValue="mcp">
        <TabsList>
          <Tooltip>
            <TooltipTrigger asChild>
              <span>
                <TabsTrigger value="mcp">MCP Server</TabsTrigger>
              </span>
            </TooltipTrigger>
            <TooltipContent>
              Recommended. Works with any MCP-compatible agent.
            </TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <span>
                <TabsTrigger value="skills">Skills</TabsTrigger>
              </span>
            </TooltipTrigger>
            <TooltipContent>
              For Claude Code, Goose, and other skill-enabled agents.
            </TooltipContent>
          </Tooltip>
        </TabsList>
        <TabsContent value="mcp">
          <div className="space-y-3 text-sm">
            <div className="space-y-2">
              <div className="flex items-center justify-between rounded-lg bg-muted px-4 py-2.5">
                <div>
                  <p className="text-xs text-muted-foreground mb-0.5">
                    Endpoint URL
                  </p>
                  <p className="font-mono text-xs">{mcpUrl}</p>
                </div>
                <Button
                  onClick={() => copyToClipboard(mcpUrl)}
                  size="sm"
                  variant="secondary"
                >
                  Copy
                </Button>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-muted px-4 py-2.5">
                <div>
                  <p className="text-xs text-muted-foreground mb-0.5">
                    Authorization Header
                  </p>
                  <p className="font-mono text-xs truncate max-w-md">
                    Bearer {connectionSecret.slice(0, 20)}...
                  </p>
                </div>
                <Button
                  onClick={() => copyToClipboard(`Bearer ${connectionSecret}`)}
                  size="sm"
                  variant="secondary"
                >
                  Copy
                </Button>
              </div>
            </div>
            <p className="text-xs text-muted-foreground">
              Set the transport to{" "}
              <span className="font-semibold">Streamable HTTP</span> and add the
              Authorization header to authenticate.
            </p>
          </div>
        </TabsContent>
        <TabsContent value="skills">
          <div className="space-y-4 text-sm">
            <div className="flex items-center gap-2 rounded-md bg-muted px-3 py-2">
              <code className="flex-1 text-xs whitespace-pre-wrap break-all">
                {` mkdir -p ~/.alby-cli && echo "${connectionSecret}" > ~/.alby-cli/connection-secret.key && npx skills add getAlby/payments-skill`}
              </code>
              <Button
                onClick={() =>
                  copyToClipboard(
                    ` mkdir -p ~/.alby-cli && echo "${connectionSecret}" > ~/.alby-cli/connection-secret.key && npx skills add getAlby/payments-skill`
                  )
                }
                size="sm"
                variant="ghost"
                className="h-6 px-2 shrink-0"
              >
                <CopyIcon className="w-3 h-3" />
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Building a Lightning app?{" "}
              <ExternalLink
                to="https://github.com/getAlby/builder-skill"
                className="underline hover:text-foreground transition-colors"
              >
                Add the Builder Skill
              </ExternalLink>
            </p>
          </div>
        </TabsContent>
      </Tabs>
    );
  }

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
              Header name: <span className="font-semibold">Authorization</span>
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

const inspirationCategories = [
  {
    label: "Payments",
    prompts: [
      "send $5 to alex@getalby.com for coffee.",
      "pay this Lightning invoice: lnbc1pjk...",
      "send 1,000 sats to each of these 20 podcast hosts.",
      "every Monday, send 50,000 sats to my savings wallet.",
    ],
  },
  {
    label: "Shopping",
    prompts: [
      "buy a $25 Netflix gift card.",
      "get me an eSIM with 5GB of data for my trip to Portugal.",
      "find me a VPN and a phone number, no KYC.",
    ],
  },
  {
    label: "Automation",
    prompts: [
      "read invoices.csv and pay all 47 Lightning invoices.",
      "check my balance every morning and message me if it drops below 100k sats.",
      "split last night's dinner bill equally between these 4 lightning addresses.",
    ],
  },
  {
    label: "Research",
    prompts: [
      "search the web for bitcoin news and summarize the top 5 stories.",
      "find the mass of the earth in mass units of mass of the sun.",
      "look up the latest nostr posts about AI agents.",
    ],
  },
  {
    label: "Builder",
    prompts: [
      "build a paywall for my API that charges 10 sats per request.",
      "scaffold a Next.js app with Lightning login using Alby JS SDK.",
      "create a tipping page that generates invoices on the fly.",
    ],
  },
];

function RotatingPrompt({ prompts }: { prompts: string[] }) {
  const [index, setIndex] = React.useState(0);
  const [fading, setFading] = React.useState(false);

  React.useEffect(() => {
    setIndex(0);
    setFading(false);
  }, [prompts]);

  React.useEffect(() => {
    const interval = setInterval(() => {
      setFading(true);
      setTimeout(() => {
        setIndex((i) => (i + 1) % prompts.length);
        setFading(false);
      }, 200);
    }, 4000);
    return () => clearInterval(interval);
  }, [prompts]);

  return (
    <div className="flex items-center gap-3 rounded-lg bg-muted/50 border border-border px-4 py-3">
      <span className="text-muted-foreground select-none">&rsaquo;</span>
      <p
        className={`flex-1 text-sm font-mono transition-opacity duration-200 ${
          fading ? "opacity-0" : "opacity-100"
        }`}
      >
        {prompts[index]}
      </p>
      <button
        onClick={() => copyToClipboard(prompts[index])}
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
    <Tabs defaultValue={inspirationCategories[0].label}>
      <div className="rounded-xl border border-border bg-card px-5 py-4 space-y-3">
        <div className="flex items-center justify-between">
          <p className="font-semibold text-sm">
            Give your agent new capabilities
          </p>
          <ExternalLink
            to="https://402index.io"
            className="text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            View all services
            <ArrowRightIcon className="w-3 h-3 inline ml-0.5" />
          </ExternalLink>
        </div>
        <TabsList>
          {inspirationCategories.map((cat) => (
            <TabsTrigger key={cat.label} value={cat.label}>
              {cat.label}
            </TabsTrigger>
          ))}
        </TabsList>
        {inspirationCategories.map((cat) => (
          <TabsContent key={cat.label} value={cat.label} className="mt-0">
            <RotatingPrompt prompts={cat.prompts} />
          </TabsContent>
        ))}
      </div>
    </Tabs>
  );
}
