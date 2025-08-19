import { HammerIcon, LightbulbIcon } from "lucide-react";
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
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function Goose() {
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const { toast } = useToast();

  const appStoreApp = suggestedApps.find((app) => app.id === "goose");
  if (!appStoreApp) {
    return null;
  }

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
            app_store_app_id: "goose",
          },
        });

        setConnectionSecret(createAppResponse.pairingUri);

        toast({ title: "Goose connection created" });
      } catch (error) {
        handleRequestError(toast, "Failed to create connection", error);
      }
      setLoading(false);
    })();
  };

  const streamableHttpLink = `https://mcp.getalby.com/mcp`;
  const streamableHttpLinkWithEncodedSecret = `${streamableHttpLink}?nwc=${encodeURIComponent(connectionSecret)}`;

  const gooseDesktopLink = `goose://extension?transport=streamable_http&url=${encodeURIComponent(streamableHttpLinkWithEncodedSecret)}&id=alby&name=Alby&description=Connect%20Goose%20to%20your%20Bitcoin%20Lightning%20Wallet`;

  return (
    <div className="grid gap-5">
      <AppDetailHeader appStoreApp={appStoreApp} contentRight={null} />
      {connectionSecret && (
        <div className="max-w-lg flex flex-col gap-5">
          <p>
            If you haven't installed Goose yet,{" "}
            <a
              href="https://block.github.io/goose/docs/getting-started/installation"
              target="_blank"
              className="underline"
            >
              click here
            </a>
            . Make sure to also configure Goose to connect to a LLM with a
            decent model (e.g. Claude Sonnet 3.7). You can use a{" "}
            <a href="https://ppq.ai/" target="_blank" className="underline">
              PPQ.ai API key
            </a>{" "}
            and buy AI credits with lightning.
          </p>
          <Accordion type="single" collapsible>
            <AccordionItem value="desktop">
              <AccordionTrigger>Goose Desktop</AccordionTrigger>
              <AccordionContent>
                <ExternalLinkButton to={gooseDesktopLink}>
                  Connect to Goose Desktop
                </ExternalLinkButton>
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="cli">
              <AccordionTrigger>Goose CLI</AccordionTrigger>
              <AccordionContent>
                <ul className="list-decimal list-inside">
                  <li className="list-item">
                    Run <span className="font-semibold">goose configure</span>
                  </li>
                  <li className="list-item">
                    Choose{" "}
                    <span className="font-semibold">
                      Remote Extension (Streaming HTTP)
                    </span>
                  </li>
                  <li className="list-item">
                    Set the name: <span className="font-semibold">Alby</span>
                  </li>
                  <li className="list-item">
                    Set the Streaming HTTP endpoint URI:{" "}
                    <Button
                      onClick={() => copyToClipboard(streamableHttpLink, toast)}
                      size="sm"
                    >
                      Copy URI
                    </Button>
                  </li>
                  <li className="list-item">
                    Set the timeout: <span className="font-semibold">300</span>
                  </li>
                  <li className="list-item">
                    Set a description: <span className="font-semibold">no</span>
                  </li>
                  <li className="list-item">
                    Would you like to add custom headers:{" "}
                    <span className="font-semibold">yes</span>
                  </li>
                  <li className="list-item">
                    Header name:{" "}
                    <span className="font-semibold">Authorization</span>
                  </li>
                  <li className="list-item">
                    Header value:{" "}
                    <Button
                      onClick={() =>
                        copyToClipboard(`Bearer ${connectionSecret}`, toast)
                      }
                      size="sm"
                    >
                      Copy value
                    </Button>
                  </li>
                  <li className="list-item">
                    Add another header:{" "}
                    <span className="font-semibold">no</span>
                  </li>
                </ul>
              </AccordionContent>
            </AccordionItem>
          </Accordion>

          <Alert>
            <LightbulbIcon />
            <AlertTitle>Enable the in-built Goose Memory extension</AlertTitle>
            <AlertDescription>
              <p>
                It acts like a contact list for your agent e.g. "My friend
                Rene's lightning address is reneaaron@getalby.com. Please save
                it to your memory."
              </p>
            </AlertDescription>
          </Alert>
          <Alert>
            <HammerIcon />
            <AlertTitle>For builders</AlertTitle>
            <AlertDescription>
              <p>
                You can use{" "}
                <a
                  href="https://github.com/getAlby/paidmcp"
                  target="_blank"
                  className="underline"
                >
                  PaidMCP
                </a>{" "}
                to create paid MCP tools that can be used by Alby MCP.
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
                By connecting Goose to your Alby Hub, you can teach your agent
                to speak bitcoin. This is possible through{" "}
                <a
                  href="https://github.com/getAlby/mcp"
                  target="_blank"
                  className="underline"
                >
                  Alby MCP
                </a>{" "}
                - an MCP server that connects to NWC wallets.
              </p>
              <div className=" flex flex-col gap-5">
                <p className="text-muted-foreground">
                  Connect your hub to goose to:
                </p>
                <ul className="text-muted-foreground pl-4 gap-2 flex flex-col">
                  <li>
                    💬 Interact with your wallet with natural language:{" "}
                    <span className="italic">"Pay $1 to my friend Rene"</span>
                  </li>
                  <li>
                    ⚡ Give Goose access to paid tools:{" "}
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
