import { copyToClipboard } from "src/lib/clipboard";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Button } from "src/components/ui/button";

export function GooseConnectionInstructions({
  gooseDesktopLink,
}: {
  gooseDesktopLink: string;
}) {
  return (
    <Accordion type="single" collapsible>
      <AccordionItem value="desktop">
        <AccordionTrigger>Goose Desktop</AccordionTrigger>
        <AccordionContent>
          <Button asChild>
            <a href={gooseDesktopLink}>Connect to Goose Desktop</a>
          </Button>
        </AccordionContent>
      </AccordionItem>
      <AccordionItem value="cli">
        <AccordionTrigger>Goose CLI</AccordionTrigger>
        <AccordionContent>
          <ol className="list-decimal list-inside space-y-1 text-sm">
            <li>
              Copy this prompt and paste it into Goose:{" "}
              <Button
                onClick={() =>
                  copyToClipboard(
                    `Install the skill from https://getalby.com/cli/SKILL.md and use the auth command to connect to my Alby Hub wallet at ${window.location.origin}`
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
