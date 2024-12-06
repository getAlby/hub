import React from "react";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { handleRequestError } from "src/utils/handleRequestError";

export function SimpleBoost() {
  const [name, setName] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);
  const { data: apps } = useApps();
  const [nwcUri, setNwcUri] = React.useState("");
  const { toast } = useToast();

  if (!apps) {
    return <Loading />;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    (async () => {
      try {
        if (apps?.some((existingApp) => existingApp.name === name)) {
          throw new Error("A connection with the same name already exists.");
        }

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
        <div className="max-w-lg flex flex-col gap-5">
          <p>Add the following widget to your website.</p>

          <div className="flex flex-col items-center relative">
            <textarea>{nwcUri}</textarea>
          </div>
        </div>
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
              <li>🔒 Supports any wallet</li>
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
