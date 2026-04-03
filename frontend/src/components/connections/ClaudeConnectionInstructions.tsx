import { copyToClipboard } from "src/lib/clipboard";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Button } from "src/components/ui/button";
import ExternalLink from "src/components/ExternalLink";

export function ClaudeConnectionInstructions({
  connectionSecret,
  mcpUrlWithSecret,
}: {
  connectionSecret: string;
  mcpUrlWithSecret: string;
}) {
  return (
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
              <ExternalLink
                to="https://claude.ai/download"
                className="underline"
              >
                Claude Desktop
              </ExternalLink>
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
              <ExternalLink
                to="https://www.anthropic.com/claude-code"
                className="underline"
              >
                Claude Code
              </ExternalLink>
            </li>
            <li>
              Copy this prompt and paste it into Claude Code:{" "}
              <Button
                onClick={() =>
                  copyToClipboard(
                    `Install the skill from https://getalby.com/cli/SKILL.md and use the setup command with this connection secret: ${connectionSecret}`
                  )
                }
                size="sm"
                variant="secondary"
              >
                Copy prompt
              </Button>
            </li>
          </ol>
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}
