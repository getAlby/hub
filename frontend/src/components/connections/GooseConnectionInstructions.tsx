import { copyToClipboard } from "src/lib/clipboard";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Button } from "src/components/ui/button";

export function GooseConnectionInstructions({
  connectionSecret,
  mcpUrl,
  gooseDesktopLink,
}: {
  connectionSecret: string;
  mcpUrl: string;
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
              Set a description: <span className="font-semibold">no</span>
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
            <li>
              Add another header: <span className="font-semibold">no</span>
            </li>
          </ol>
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}
