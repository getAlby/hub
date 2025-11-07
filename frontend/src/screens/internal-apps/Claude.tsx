import {
  ChartLineIcon,
  GiftIcon,
  LightbulbIcon,
  MessageCircleIcon,
} from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { AppDetailConnectedApps } from "src/components/connections/AppDetailConnectedApps";
import { AppStoreDetailHeader } from "src/components/connections/AppStoreDetailHeader";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import ExternalLink from "src/components/ExternalLink";
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
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function Claude() {
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");

  const appStoreApp = appStoreApps.find((app) => app.id === "claude");
  if (!appStoreApp) {
    return null;
  }

  const mcpLinkWithEncodedSecret = `https://mcp.getalby.com/mcp?nwc=${encodeURIComponent(connectionSecret)}`;

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        const createAppResponse = await createApp({
          name: appStoreApp.title,
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
          maxAmount: 10_000,
          budgetRenewal: "monthly",
          metadata: {
            app_store_app_id: "claude",
          },
        });

        setConnectionSecret(createAppResponse.pairingUri);

        toast("Claude connection created");
      } catch (error) {
        handleRequestError("Failed to create connection", error);
      }
      setLoading(false);
    })();
  };

  return (
    <div className="grid gap-5">
      <AppStoreDetailHeader appStoreApp={appStoreApp} contentRight={null} />
      {connectionSecret && (
        <div className="max-w-lg flex flex-col gap-5">
          <p>
            Click one of the below options to connect Claude to your Alby Hub.
          </p>
          <Accordion type="single" collapsible>
            <AccordionItem value="web">
              <AccordionTrigger>Claude Web</AccordionTrigger>
              <AccordionContent>
                <ul className="list-decimal list-inside">
                  <li className="list-item">
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
                  <li className="list-item">Go to Settings → Integrations</li>
                  <li className="list-item">Click +Add integration</li>
                  <li className="list-item">Integration Name: Alby</li>
                  <li className="list-item">
                    Paste the integration URL:{" "}
                    <Button
                      onClick={() => copyToClipboard(mcpLinkWithEncodedSecret)}
                      size="sm"
                    >
                      Copy URL
                    </Button>
                  </li>
                </ul>
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="desktop">
              <AccordionTrigger>Claude Desktop</AccordionTrigger>
              <AccordionContent>
                <ul className="list-decimal list-inside">
                  <li className="list-item">
                    Download and install Claude Desktop from{" "}
                    <a
                      href="https://claude.ai/download"
                      target="_blank"
                      className="underline"
                    >
                      claude.ai/download
                    </a>
                  </li>
                  <li className="list-item">Open Claude Desktop and sign in</li>
                  <li className="list-item">Go to Settings → Integrations</li>
                  <li className="list-item">Click +Add integration</li>
                  <li className="list-item">Integration Name: Alby</li>
                  <li className="list-item">
                    Paste the integration URL:{" "}
                    <Button
                      onClick={() => copyToClipboard(mcpLinkWithEncodedSecret)}
                      size="sm"
                    >
                      Copy URL
                    </Button>
                  </li>
                </ul>
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="code">
              <AccordionTrigger>Claude Code</AccordionTrigger>
              <AccordionContent>
                <ul className="list-decimal list-inside">
                  <li className="list-item">
                    Install{" "}
                    <a
                      href="https://www.anthropic.com/claude-code"
                      target="_blank"
                      className="underline"
                    >
                      Claude Code
                    </a>
                  </li>
                  <li className="list-item">
                    Paste the MCP add command into your terminal:{" "}
                    <Button
                      onClick={() =>
                        copyToClipboard(
                          `claude mcp add --transport http alby https://mcp.getalby.com/mcp --header "Authorization: Bearer ${connectionSecret}"`
                        )
                      }
                      size="sm"
                    >
                      Copy command
                    </Button>
                  </li>
                </ul>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </div>
      )}
      {!connectionSecret && (
        <>
          <Card className="max-w-lg">
            <CardHeader>
              <CardTitle className="text-2xl">About the App</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              <p className="text-muted-foreground">
                By connecting Claude to your Alby Hub, you can enable your AI
                assistant to interact with Bitcoin Lightning Network. This
                allows Claude to help you with payments, balance checks, and
                more.
              </p>
              <div className=" flex flex-col gap-5">
                <p className="text-muted-foreground">
                  Connect your hub to Claude to:
                </p>
                <ul className="text-muted-foreground pl-4 gap-4 flex flex-col">
                  <li>
                    <MessageCircleIcon className="size-4 inline" /> Interact
                    with your wallet with natural language:{" "}
                    <span className="italic">"Pay $1 to my friend Rene"</span>{" "}
                    (with{" "}
                    <ExternalLink
                      to="https://support.claude.com/en/articles/11817273-using-claude-s-chat-search-and-memory-to-build-on-previous-context"
                      className="underline"
                    >
                      Claude's Memory
                    </ExternalLink>
                    )
                  </li>
                  <li>
                    <GiftIcon className="size-4 inline" /> Buy giftcards:{" "}
                    <span className="italic">
                      "Buy a $15 doordash giftcard"
                    </span>{" "}
                    (with{" "}
                    <ExternalLink
                      to="https://www.bitrefill.com/account/developers/mcp-server"
                      className="underline"
                    >
                      Bitrefill MCP
                    </ExternalLink>
                    )
                  </li>
                  <li>
                    <ChartLineIcon className="size-4 inline" /> Let Claude trade
                    for you:{" "}
                    <span className="italic">
                      "Open a $10 2x long position and take profit at 50% higher
                      than the current price"
                    </span>{" "}
                    (with{" "}
                    <ExternalLink
                      to="https://sup3r.cool/ln-markets/"
                      className="underline"
                    >
                      LNMarkets MCP
                    </ExternalLink>
                    )
                  </li>
                  <li>
                    <LightbulbIcon className="size-4 inline" /> Use other
                    awesome paid MCP tools: (see more{" "}
                    <ExternalLink
                      to="https://github.com/getAlby/awesome-ai-bitcoin/?tab=readme-ov-file#mcp-servers"
                      className="underline"
                    >
                      Awesome MCP Servers
                    </ExternalLink>
                    )
                  </li>
                </ul>

                <form
                  onSubmit={handleSubmit}
                  className="flex flex-col items-start gap-5 max-w-lg"
                >
                  <LoadingButton loading={isLoading} type="submit">
                    Create Connection
                  </LoadingButton>
                </form>
              </div>
            </CardContent>
          </Card>
          <AppDetailConnectedApps appStoreApp={appStoreApp} showTitle />
        </>
      )}
    </div>
  );
}
