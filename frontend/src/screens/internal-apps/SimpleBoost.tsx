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
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { Textarea } from "src/components/ui/textarea";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function SimpleBoost() {
  const [name, setName] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const [nwcUri, setNwcUri] = React.useState("");
  const [scriptContent, setScriptContent] = React.useState("");
  const { toast } = useToast();

  React.useEffect(() => {
    if (nwcUri) {
      setScriptContent(`<script type="module" src="https://esm.sh/simple-boost@latest"></script>
<simple-boost currency="usd" amount="1.0" nwc="${nwcUri}">Boost $1.00</simple-boost>`);
    }
  }, [nwcUri]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        const createAppResponse = await createApp({
          name,
          scopes: ["lookup_invoice", "make_invoice"],
          isolated: true,
          metadata: {
            app_store_app_id: "simpleboost",
          },
        });

        setNwcUri(createAppResponse.pairingUri);

        toast({ title: "Simple Boost connection created" });
      } catch (error) {
        handleRequestError(toast, "Failed to create connection", error);
      }
      setLoading(false);
    })();
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="SimpleBoost"
        description="The donation button for your website."
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
                <CopyIcon />
                Copy
              </Button>
            </div>
            <p className="text-sm text-muted-foreground mt-3">
              By default the SimpleBoost widget is loaded from a CDN. See other
              options{" "}
              <ExternalLink
                to="https://getalby.github.io/simple-boost/install/"
                className="font-medium text-foreground underline"
              >
                here
              </ExternalLink>
            </p>
          </CardContent>
        </Card>
      )}
      {!nwcUri && (
        <>
          <div className="max-w-lg flex flex-col gap-5">
            <p className="text-muted-foreground">
              SimpleBoost is a donation button for your website. Add the widget
              to your website and allow your visitors to send sats with the
              click of a button.
            </p>
            <ul className="text-muted-foreground">
              <li>⚡ Lightning fast transactions directly to your Alby Hub</li>
              <li>🔗 Set amounts in Bitcoin or any other currency</li>
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
