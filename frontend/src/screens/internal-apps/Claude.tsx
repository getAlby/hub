import React from "react";
import { AppDetailConnectedApps } from "src/components/connections/AppDetailConnectedApps";
import { AppDetailHeader } from "src/components/connections/AppDetailHeader";
import { suggestedApps } from "src/components/connections/SuggestedAppData";
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
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function Claude() {
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const { toast } = useToast();

  const appStoreApp = suggestedApps.find((app) => app.id === "claude");
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

        toast({ title: "Claude connection created" });
      } catch (error) {
        handleRequestError(toast, "Failed to create connection", error);
      }
      setLoading(false);
    })();
  };

  return (
    <div className="grid gap-5">
      <AppDetailHeader appStoreApp={appStoreApp} contentRight={null} />
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
                      onClick={() =>
                        copyToClipboard(mcpLinkWithEncodedSecret, toast)
                      }
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
                      onClick={() =>
                        copyToClipboard(mcpLinkWithEncodedSecret, toast)
                      }
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
                          `claude mcp add --transport http alby https://mcp.getalby.com/mcp --header "Authorization: Bearer ${connectionSecret}"`,
                          toast
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
                <ul className="text-muted-foreground pl-4 gap-2 flex flex-col">
                  <li>
                    💬 Interact with your wallet with natural language:{" "}
                    <span className="italic">"Pay $1 to my friend Rene"</span>
                  </li>
                  <li>
                    ⚡ Give Claude access to paid tools:{" "}
                    <span className="italic">
                      "Buy a $15 doordash giftcard"
                    </span>
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
