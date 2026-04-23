import {
  ArrowRightIcon,
  ArrowUpRightIcon,
  BotIcon,
  BoxIcon,
  CheckCircleIcon,
  ChevronRightIcon,
  CopyIcon,
  EyeOffIcon,
  HammerIcon,
  InfoIcon,
  LayersIcon,
  LayoutGridIcon,
  type LucideIcon,
  RepeatIcon,
  ShieldCheckIcon,
  ShoppingBagIcon,
  SparklesIcon,
  XIcon,
  ZapIcon,
} from "lucide-react";
import React from "react";
import { Link, useNavigate } from "react-router";
import { toast } from "sonner";
import bitrefillLogo from "src/assets/suggested-apps/bitrefill.png";
import claudeLogo from "src/assets/suggested-apps/claude.png";
import clineLogo from "src/assets/suggested-apps/cline.png";
import codexLogo from "src/assets/suggested-apps/codex.png";
import cursorLogo from "src/assets/suggested-apps/cursor.png";
import geminiLogo from "src/assets/suggested-apps/gemini.png";
import gooseLogo from "src/assets/suggested-apps/goose.png";
import openclawLogo from "src/assets/suggested-apps/openclaw.png";
import opencodeLogo from "src/assets/suggested-apps/opencode.png";
import payperqLogo from "src/assets/suggested-apps/payperq.png";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
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
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

type Agent = {
  id: string;
  name: string;
  logo: string;
  description: string;
  setupUrl: string;
  mcpInstructions?: () => React.ReactNode;
};

const agents: Agent[] = [
  {
    id: "openclaw",
    name: "OpenClaw",
    logo: openclawLogo,
    description: "Open-source personal AI assistant",
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
    description: "Anthropic's AI assistant (Code, Web & Desktop)",
    setupUrl: "",
    mcpInstructions: () => (
      <ol className="list-decimal list-inside space-y-1 text-sm">
        <li>
          Open{" "}
          <a
            href="https://claude.ai"
            target="_blank"
            className="underline font-medium"
          >
            claude.ai
          </a>{" "}
          or Claude Desktop and sign in
        </li>
        <li>Go to Settings &rarr; Connectors</li>
        <li>Add custom connector</li>
        <li>Enter &ldquo;Alby&rdquo; as connector name</li>
        <li>Paste the URL below as the connector URL</li>
      </ol>
    ),
  },
  {
    id: "gemini",
    name: "Gemini CLI",
    logo: geminiLogo,
    description: "Google's open-source AI agent for the terminal",
    setupUrl: "",
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
    setupUrl: "",
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
  const [selectedAgent, setSelectedAgent] = React.useState<string | null>(null);
  const [heroDismissed, setHeroDismissed] = React.useState(
    () => localStorage.getItem(localStorageKeys.aiHeroDismissed) === "true"
  );

  const dismissHero = React.useCallback(() => {
    setHeroDismissed(true);
    localStorage.setItem(localStorageKeys.aiHeroDismissed, "true");
  }, []);

  const selectedAgentData = agents.find((a) => a.id === selectedAgent);
  const showAuthPrompt = selectedAgent && !selectedAgentData?.setupUrl;

  const handleConnect = () => {
    if (selectedAgentData?.setupUrl) {
      navigate(selectedAgentData.setupUrl);
    }
  };

  return (
    <>
      <AppHeader
        title="AI & Agents"
        pageTitle="AI & Agents"
        contentRight={
          <LinkButton to="/apps?tab=app-store&category=ai" variant="outline">
            <LayoutGridIcon className="w-4 h-4" />
            Explore App Store
          </LinkButton>
        }
      />

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
                  AI + Bitcoin
                </p>
                <h2 className="text-4xl lg:text-5xl font-bold tracking-tight leading-[1.1] mb-4">
                  Give your AI agent
                  <br />
                  <span className="text-primary">a wallet</span>
                </h2>
                <p className="text-muted-foreground text-lg max-w-md">
                  Connect any AI agent to your Alby Hub and let it send
                  payments, buy gift cards, and access paid services on its own.
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
                      <span className="w-2 h-4 bg-primary ml-0.5 animate-[blink_1s_step-end_3]" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Why Lightning — value props */}
            <div className="border-t border-border grid grid-cols-1 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x divide-border">
              {whyLightningItems.map((item) => {
                const Icon = item.icon;
                return (
                  <div key={item.title} className="p-6 lg:p-8">
                    <Icon className="w-5 h-5 text-primary mb-2" />
                    <h3 className="font-semibold text-lg mb-1">{item.title}</h3>
                    <p className="text-muted-foreground text-sm">
                      {item.description}
                    </p>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Connect section */}
      <div className="space-y-4">
        <Card className="border-primary/75 bg-primary/5">
          <CardHeader>
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                {selectedAgentData ? (
                  <img
                    src={selectedAgentData.logo}
                    alt={selectedAgentData.name}
                    className="w-10 h-10 rounded-lg shrink-0"
                  />
                ) : (
                  <div className="w-10 h-10 rounded-lg bg-primary/20 flex items-center justify-center shrink-0">
                    <BotIcon className="w-5 h-5 text-primary" />
                  </div>
                )}
                <div>
                  <CardTitle>Connect Your Agent</CardTitle>
                  <p className="text-sm text-muted-foreground mt-1">
                    Pick your agent, create a connection, and follow the setup
                    steps
                  </p>
                </div>
              </div>
              <Link
                to="/apps?tab=connected-apps"
                className="text-xs text-muted-foreground hover:text-foreground transition-colors flex items-center gap-1"
              >
                Manage Connections
                <ArrowRightIcon className="w-3 h-3" />
              </Link>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-2">
              <Select
                value={selectedAgent ?? undefined}
                onValueChange={(value) => {
                  setSelectedAgent(value);
                }}
              >
                <SelectTrigger className="w-60">
                  <SelectValue
                    placeholder={
                      <span className="flex items-center gap-2">
                        <span className="flex -space-x-2">
                          {agents.slice(0, 4).map((agent) => (
                            <img
                              key={agent.id}
                              src={agent.logo}
                              alt={agent.name}
                              className="w-5 h-5 rounded border-2 border-background"
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
              {selectedAgentData?.setupUrl && (
                <Button onClick={handleConnect}>
                  <ZapIcon className="w-4 h-4" />
                  Connect
                </Button>
              )}
            </div>

            {/* Auth prompt shown immediately for CLI agents */}
            {showAuthPrompt && (
              <GenericAuthPrompt
                agent={
                  selectedAgentData ?? {
                    id: "other",
                    name: "your agent",
                    logo: "",
                    description: "",
                    setupUrl: "",
                  }
                }
              />
            )}
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
                        1,000+ APIs
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

function GenericAuthPrompt({ agent }: { agent: Agent }) {
  const [mcpMode, setMcpMode] = React.useState(false);
  const basePath = import.meta.env.BASE_URL.replace(/\/$/, "");
  const hubUrl = window.location.origin + basePath;
  const genericPrompt = `Run \`npx -y skills add getAlby/payments-skill -y\` and connect to my Alby Hub wallet at ${hubUrl}`;

  return (
    <div
      key={agent.id}
      className="space-y-3 text-sm animate-[flash_0.4s_ease-out]"
    >
      {mcpMode && agent.mcpInstructions ? (
        <McpSetup
          agentName={agent.name}
          mcpInstructions={agent.mcpInstructions}
          onBack={() => setMcpMode(false)}
        />
      ) : (
        <>
          <p className="text-muted-foreground">
            Copy this prompt and paste it into {agent.name}:
          </p>
          <button
            onClick={() => copyToClipboard(genericPrompt)}
            className="flex items-center gap-3 rounded-lg bg-muted/50 border border-border px-4 py-3 w-full text-left cursor-pointer hover:border-primary/30 transition-colors"
          >
            <ChevronRightIcon className="w-3.5 h-3.5 text-muted-foreground shrink-0 self-start mt-0.5" />
            <p className="flex-1 text-sm font-mono break-all select-none">
              {genericPrompt}
            </p>
            <CopyIcon className="w-4 h-4 text-muted-foreground shrink-0" />
          </button>
          {agent.mcpInstructions && (
            <p className="text-xs text-muted-foreground">
              <button
                onClick={() => setMcpMode(true)}
                className="underline hover:text-foreground transition-colors"
              >
                Using {agent.name} Web / Desktop?
              </button>
            </p>
          )}
        </>
      )}
    </div>
  );
}

function McpSetup({
  agentName,
  mcpInstructions,
  onBack,
}: {
  agentName: string;
  mcpInstructions: () => React.ReactNode;
  onBack: () => void;
}) {
  const [isLoading, setIsLoading] = React.useState(false);
  const [mcpUrl, setMcpUrl] = React.useState("");
  const [createdAppId, setCreatedAppId] = React.useState<number>();

  const handleCreateConnection = React.useCallback(async () => {
    if (isLoading || mcpUrl) {
      return;
    }
    setIsLoading(true);
    try {
      const response = await createApp({
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
          app_store_app_id: agentName.toLowerCase().replace(/\s+/g, "-"),
        },
      });
      setMcpUrl(
        `https://mcp.getalby.com/mcp?nwc=${encodeURIComponent(response.pairingUri)}`
      );
      setCreatedAppId(response.id);
      toast("Connection created");
    } catch (error) {
      handleRequestError("Failed to create connection", error);
    }
    setIsLoading(false);
  }, [agentName, isLoading, mcpUrl]);

  // Create connection immediately when the component mounts
  React.useEffect(() => {
    handleCreateConnection();
  }, [handleCreateConnection]);

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Loading className="size-4" />
        Creating connection...
      </div>
    );
  }

  if (!mcpUrl) {
    return (
      <Button onClick={handleCreateConnection} size="sm">
        Retry
      </Button>
    );
  }

  const maskedUrl = `https://mcp.getalby.com/mcp?nwc=${"•".repeat(20)}`;

  return (
    <div className="space-y-3 text-sm">
      {mcpInstructions()}
      <button
        onClick={() => copyToClipboard(mcpUrl)}
        className="flex items-center gap-2 rounded-lg bg-muted/50 border border-border px-3 py-2 w-full text-left cursor-pointer hover:border-primary/30 transition-colors"
      >
        <p className="flex-1 text-xs font-mono select-none truncate">
          {maskedUrl}
        </p>
        <CopyIcon className="w-4 h-4 text-muted-foreground shrink-0" />
      </button>
      <div className="flex items-center gap-3">
        {createdAppId && (
          <Link
            to={`/apps/${createdAppId}`}
            className="text-xs text-muted-foreground hover:text-foreground underline transition-colors"
          >
            Edit budget & permissions
          </Link>
        )}
        <button
          onClick={onBack}
          className="text-xs text-muted-foreground hover:text-foreground underline transition-colors"
        >
          Use CLI instead
        </button>
      </div>
    </div>
  );
}

const whyLightningItems = [
  {
    icon: ShieldCheckIcon,
    title: "Stay in Control",
    description:
      "Set spending limits per agent. You decide how much it can spend and when budgets reset — no surprise bills.",
  },
  {
    icon: EyeOffIcon,
    title: "Private by Default",
    description:
      "No credit cards or personal info shared with merchants. Your agent pays over Lightning — fast, direct, and private.",
  },
  {
    icon: ZapIcon,
    title: "Instant Access to Paid Services",
    description:
      "Your agent can access 1,000+ paid APIs instantly — gift cards, domains, hosting, AI models, and more.",
  },
];

const inspirationCategories: {
  label: string;
  icon: LucideIcon;
  prompts: string[];
  skill?: { prompt: string; skillName: string; url: string };
}[] = [
  {
    label: "Wallet",
    icon: ZapIcon,
    prompts: [
      "send $5 to hub@getalby.com for coffee",
      "how much is $10 in sats right now?",
      "make an invoice for 50,000 sats",
    ],
  },
  {
    label: "Shopping",
    icon: ShoppingBagIcon,
    prompts: [
      "buy a $25 Netflix gift card on bitrefill.com",
      "get me an eSIM with 5GB of data for my trip to Portugal on bitrefill.com",
      "what gift cards are available in the US on bitrefill.com?",
    ],
    skill: {
      prompt:
        "Run `npx -y skills add bitrefill/agents -y` to install the Bitrefill Skill",
      skillName: "Bitrefill Skill",
      url: "https://bitrefill.com/agents",
    },
  },
  {
    label: "Creative",
    icon: SparklesIcon,
    prompts: [
      "generate a watercolor painting of a mountain cabin at sunset on ppq.ai",
      "generate a cool bitcoin logo on ppq.ai and print it on a t-shirt on unhuman.store",
      "create a logo for my coffee shop using ppq.ai image generation",
    ],
  },
  {
    label: "Services",
    icon: LayersIcon,
    prompts: [
      "search podcasts for discussions about bitcoin scaling",
      "set up an anonymous email address on lnemail.net",
      "buy the domain my-awesome-project.dev on unhuman.domains",
      "spin up a VPS with 2 cores and 4GB RAM on lnvps.net",
    ],
  },
  {
    label: "Automation",
    icon: RepeatIcon,
    prompts: [
      "read payouts.csv and send 1,000 sats to each lightning address",
      "calculate how much I spent this month and break it down by day",
      "export all my transactions from the last 12 months as a CSV",
    ],
  },
  {
    label: "Build Apps",
    icon: HammerIcon,
    prompts: [
      "build an AI image generator that charges 500 sats per image",
      "create a file converter that charges 50 sats per conversion",
      "build a blog where readers unlock articles for 50 sats each",
    ],
    skill: {
      prompt:
        "Run `npx -y skills add getAlby/builder-skill -y` to install the Builder Skill",
      skillName: "Builder Skill",
      url: "https://github.com/getAlby/builder-skill",
    },
  },
  {
    label: "Node",
    icon: BoxIcon,
    prompts: [
      "open a channel with 2M sats to ACINQ's node",
      "show me my channels and their balances",
      "what's my node's connection info?",
    ],
    skill: {
      prompt:
        "Run `npx -y skills add getAlby/hub-skill -y` to install the Alby Hub Skill",
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
    <button
      onClick={() => copyToClipboard(currentPrompt)}
      className="flex items-center gap-3 rounded-lg bg-muted/50 border border-border px-4 py-3 w-full text-left cursor-pointer hover:border-primary/30 transition-colors"
    >
      <ChevronRightIcon className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
      <p className="flex-1 text-sm font-mono">
        {currentPrompt.slice(0, charCount)}
        <span
          className={`inline-block w-2 h-[1.1em] translate-y-[2px] ml-0.5 bg-primary ${
            isTyping ? "" : "animate-[blink_1s_step-end_infinite]"
          }`}
        />
      </p>
      <CopyIcon className="w-4 h-4 text-muted-foreground shrink-0" />
    </button>
  );
}

function InspirationPrompts() {
  return (
    <Tabs defaultValue={inspirationCategories[0].label}>
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-5 pt-5 pb-4">
          <p className="font-semibold text-sm mb-4">What can your agent do?</p>
          <TabsList variant="line">
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
