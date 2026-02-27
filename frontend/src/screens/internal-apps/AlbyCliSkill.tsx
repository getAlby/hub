import { BotIcon, CopyIcon, ShieldCheckIcon, WalletIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { AppDetailConnectedApps } from "src/components/connections/AppDetailConnectedApps";
import { AppStoreDetailHeader } from "src/components/connections/AppStoreDetailHeader";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
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

const skillInstallCommand = "npx skills add getAlby/alby-cli-skill";
const skillInstallPrompt =
  "Install this skill as a custom skill: https://raw.githubusercontent.com/getAlby/alby-cli-skill/refs/heads/master/SKILL.md";
const verifyPrompt = "What's your wallet balance?";

export function AlbyCliSkill() {
  const [isLoading, setLoading] = React.useState(false);
  const [connectionSecret, setConnectionSecret] = React.useState("");

  const appStoreApp = appStoreApps.find((app) => app.id === "alby-cli-skill");
  if (!appStoreApp) {
    return null;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);

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
          app_store_app_id: "alby-cli-skill",
        },
      });

      setConnectionSecret(createAppResponse.pairingUri);
      toast("Alby CLI Skill connection created");
    } catch (error) {
      handleRequestError("Failed to create connection", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid gap-5">
      <AppStoreDetailHeader appStoreApp={appStoreApp} contentRight={null} />

      {connectionSecret ? (
        <div className="max-w-3xl flex flex-col gap-5">
          <p className="text-muted-foreground">
            Follow these steps to install and use Alby CLI Skill with your AI
            agent.
          </p>

          <Accordion type="single" collapsible>
            <AccordionItem value="install">
              <AccordionTrigger>Install with one command</AccordionTrigger>
              <AccordionContent>
                <div className="flex items-center gap-2">
                  <div className="font-mono text-foreground text-sm break-all bg-muted p-2 rounded">
                    {skillInstallCommand}
                  </div>
                  <Button
                    size="sm"
                    onClick={() => copyToClipboard(skillInstallCommand)}
                  >
                    <CopyIcon />
                    Copy command
                  </Button>
                </div>
              </AccordionContent>
            </AccordionItem>

            <AccordionItem value="openclaw">
              <AccordionTrigger>OpenClaw setup</AccordionTrigger>
              <AccordionContent>
                <ul className="list-decimal list-inside space-y-2 text-muted-foreground">
                  <li>
                    Tell your agent to install this skill as a custom skill:
                    <div className="my-4 flex items-center gap-4">
                      <div className="font-mono text-foreground text-sm break-all bg-muted p-2 rounded">
                        {skillInstallPrompt}
                      </div>
                      <Button
                        size="sm"
                        onClick={() => copyToClipboard(skillInstallPrompt)}
                      >
                        <CopyIcon />
                        Copy Prompt
                      </Button>
                    </div>
                  </li>
                  <li>
                    Save your wallet connection secret at
                    <code className="ml-1 text-foreground">
                      ~/.alby-cli/connection-secret.key
                    </code>
                    .
                  </li>
                </ul>

                <div className="mt-4 flex flex-col gap-2">
                  <div className="text-sm text-muted-foreground">
                    Connection secret
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      onClick={() => copyToClipboard(connectionSecret)}
                    >
                      <CopyIcon />
                      Copy Connection secret
                    </Button>
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>

            <AccordionItem value="verify">
              <AccordionTrigger>Verify it works</AccordionTrigger>
              <AccordionContent>
                <p className="text-muted-foreground">
                  Ask your agent this question:
                </p>
                <div className="mt-2 flex items-center gap-2">
                  <div className="font-mono text-foreground text-sm break-all bg-muted p-2 rounded">
                    {verifyPrompt}
                  </div>
                  <Button
                    size="sm"
                    onClick={() => copyToClipboard(verifyPrompt)}
                  >
                    <CopyIcon />
                    Copy prompt
                  </Button>
                </div>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </div>
      ) : (
        <>
          <Card className="max-w-lg">
            <CardHeader>
              <CardTitle className="text-2xl">About the App</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-5">
              <p className="text-muted-foreground">
                Alby CLI Skill is built for AI agents that need to use Alby CLI
                more effectively. It connects your Alby Hub to skill-enabled
                agents so they can use Bitcoin payments inside autonomous
                workflows.
              </p>

              <ul className="text-muted-foreground pl-4 flex flex-col gap-3">
                <li>
                  <WalletIcon className="size-4 inline" /> Give your AI agent a
                  programmable Lightning wallet to check balances, create
                  invoices, and send payments from prompts.
                </li>
                <li>
                  <BotIcon className="size-4 inline" /> Enable agentic payment
                  workflows in OpenClaw with wallet context for multi-step
                  tasks.
                </li>
                <li>
                  <ShieldCheckIcon className="size-4 inline" /> Use scoped NWC
                  permissions with budgets so you can grant access safely and
                  stay in control of spending.
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
            </CardContent>
          </Card>
          <AppDetailConnectedApps appStoreApp={appStoreApp} showTitle />
        </>
      )}
    </div>
  );
}
