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
import {
  DEFAULT_APP_BUDGET_RENEWAL,
  DEFAULT_APP_BUDGET_SATS,
} from "src/constants";
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

  const mcpUrlWithSecret = `https://mcp.getalby.com/mcp?nwc=${encodeURIComponent(connectionSecret)}`;

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
          maxAmount: DEFAULT_APP_BUDGET_SATS,
          budgetRenewal: DEFAULT_APP_BUDGET_RENEWAL,
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
                <ol className="list-decimal list-inside space-y-1 text-sm">
                  <li>
                    Visit{" "}
                    <ExternalLink to="https://claude.ai" className="underline">
                      claude.ai
                    </ExternalLink>{" "}
                    and sign in
                  </li>
                  <li>Go to Settings &rarr; Connectors</li>
                  <li>Add custom connector</li>
                  <li>Enter "Alby" as connector name</li>
                  <li>
                    Paste the MCP server URL{" "}
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
                    <ExternalLink
                      to="https://claude.ai/download"
                      className="underline"
                    >
                      Claude Desktop
                    </ExternalLink>
                  </li>
                  <li>Open Claude Desktop and sign in</li>
                  <li>Go to Settings &rarr; Connectors</li>
                  <li>Add custom connector</li>
                  <li>Enter "Alby" as connector name</li>
                  <li>
                    Paste the MCP server URL{" "}
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
                      "Analyze market sentiment and trading data from the past 3
                      months and based on this, open a $10 2x long or short
                      position"
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
