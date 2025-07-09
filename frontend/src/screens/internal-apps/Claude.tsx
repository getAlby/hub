import { LightbulbIcon } from "lucide-react";
import React from "react";
import AppHeader from "src/components/AppHeader";
import { suggestedApps } from "src/components/SuggestedAppData";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function Claude() {
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const { toast } = useToast();

  const app = suggestedApps.find((app) => app.id === "claude");
  if (!app) {
    return null;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        const createAppResponse = await createApp({
          name: app.title,
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
      <AppHeader
        title="Claude"
        description="AI assistant for conversations, analysis, and coding."
      />
      {connectionSecret && (
        <div className="max-w-lg flex flex-col gap-5">
          <p>
            You can connect Claude to your Alby Hub by manually pasting the
            connection URI into Claude's settings. No automatic linking is
            supported.
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
                  <li className="list-item">
                    Find the wallet connection section
                  </li>
                  <li className="list-item">
                    Paste the connection URI:{" "}
                    <Button
                      onClick={() => copyToClipboard(connectionSecret, toast)}
                      size="sm"
                    >
                      Copy URI
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
                  <li className="list-item">
                    Find the wallet connection section
                  </li>
                  <li className="list-item">
                    Paste the connection URI:{" "}
                    <Button
                      onClick={() => copyToClipboard(connectionSecret, toast)}
                      size="sm"
                    >
                      Copy URI
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
                    Install the Claude extension in your code editor
                  </li>
                  <li className="list-item">Open the extension settings</li>
                  <li className="list-item">
                    Find the wallet connection configuration
                  </li>
                  <li className="list-item">
                    Paste the connection URI:{" "}
                    <Button
                      onClick={() => copyToClipboard(connectionSecret, toast)}
                      size="sm"
                    >
                      Copy URI
                    </Button>
                  </li>
                </ul>
              </AccordionContent>
            </AccordionItem>
          </Accordion>

          <Alert>
            <LightbulbIcon />
            <AlertTitle>Manual Configuration Required</AlertTitle>
            <AlertDescription>
              <p>
                Unlike some other apps, Claude requires manual configuration.
                You'll need to copy and paste the connection URI into Claude's
                settings yourself.
              </p>
            </AlertDescription>
          </Alert>
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
                    💬 Ask Claude about your wallet balance:{" "}
                    <span className="italic">"What's my current balance?"</span>
                  </li>
                  <li>
                    ⚡ Get help with Lightning payments:{" "}
                    <span className="italic">
                      "Help me send 1000 sats to this invoice"
                    </span>
                  </li>
                  <li>
                    📊 Analyze your transaction history:{" "}
                    <span className="italic">
                      "Show me my recent transactions"
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
        </>
      )}
    </div>
  );
}
