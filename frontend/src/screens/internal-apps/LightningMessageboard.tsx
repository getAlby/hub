import { CopyIcon } from "lucide-react";
import React from "react";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Textarea } from "src/components/ui/textarea";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function LightningMessageboard() {
  const [name, setName] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const [nwcUri, setNwcUri] = React.useState("");
  const [scriptContent, setScriptContent] = React.useState("");
  const { toast } = useToast();

  React.useEffect(() => {
    if (nwcUri) {
      setScriptContent(`<script type="module" src="https://esm.sh/@getalby/lightning-messageboard@latest"></script>
<lightning-messageboard nwc-url="${nwcUri}"></lightning-messageboard>`);
    }
  }, [nwcUri]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        const createAppResponse = await createApp({
          name,
          scopes: ["lookup_invoice", "make_invoice", "list_transactions"],
          isolated: true,
          metadata: {
            app_store_app_id: "lightning-messageboard",
          },
        });

        setNwcUri(createAppResponse.pairingUri);

        toast({ title: "Lightning Messageboard connection created" });
      } catch (error) {
        handleRequestError(toast, "Failed to create connection", error);
      }
      setLoading(false);
    })();
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Lightning Messageboard"
        description="A paid message board for your website."
      />
      {nwcUri && (
        <Card>
          <CardHeader>
            <CardTitle>How to Add Widget</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col">
            <p className="text-foreground">
              Paste the following code into an HTML block on your website.
            </p>
            <div className="flex gap-2 mt-4">
              <Textarea
                className="h-36 font-mono"
                value={scriptContent}
                onChange={(e) => setScriptContent(e.target.value)}
              />
              <Button
                onClick={() => copyToClipboard(scriptContent, toast)}
                variant="outline"
              >
                <CopyIcon className="w-4 h-4 mr-2" />
                Copy
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
      {!nwcUri && (
        <>
          <div className="max-w-lg flex flex-col gap-5">
            <p className="text-muted-foreground">
              Lightning Messageboard is a simple messageboard widget for your
              website. Add the widget to your website and allow your visitors to
              pay to send messages.{" "}
              <ExternalLink
                to="https://getalby.github.io/lightning-messageboard/demo.html"
                className="underline"
              >
                See Demo
              </ExternalLink>
            </p>
            <ul className="text-muted-foreground">
              <li>⚡ Lightning fast transactions directly to your Alby Hub</li>
              <li>
                🎮 Visitors can pay more to ensure their comment shows at the
                top of the message board
              </li>
              <li>
                🔒 No lightning address required, secure read-only connection
              </li>
              <li>🤙 Can be paid from any lightning wallet</li>
            </ul>
            <form
              onSubmit={handleSubmit}
              className="flex flex-col items-start gap-5 max-w-lg"
            >
              <div className="w-full grid gap-1.5">
                <Label htmlFor="name">Website name</Label>
                <Input
                  autoFocus
                  type="text"
                  name="name"
                  value={name}
                  id="name"
                  onChange={(e) => setName(e.target.value)}
                  required
                  autoComplete="off"
                />
              </div>
              <LoadingButton loading={isLoading} type="submit">
                Next
              </LoadingButton>
            </form>
          </div>
        </>
      )}
    </div>
  );
}
