import { Rocket, TriangleAlert } from "lucide-react";
import React from "react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import { Textarea } from "src/components/ui/textarea";
import { getAuthToken } from "src/lib/auth";

export default function DeveloperSettings() {
  const [show, setShowToken] = React.useState(false);
  return (
    <>
      <Alert>
        <Rocket className="h-4 w-4" />
        <AlertTitle>Power your Apps with NWC</AlertTitle>
        <AlertDescription>
          <div className="mb-2">
            Alby Hub can power your lightning apps and services in all
            environments using the amazing open protocol{" "}
            <span className="font-semibold">Nostr Wallet Connect</span>.
          </div>
          <ExternalLinkButton size={"sm"} to="https://nwc.dev">
            Learn More
          </ExternalLinkButton>
        </AlertDescription>
      </Alert>
      <Alert>
        <TriangleAlert className="h-4 w-4" />
        <AlertTitle>Experimental</AlertTitle>
        <AlertDescription>
          <div className="mb-2">
            You can use your auth token in apps to access the Alby Hub internal
            API.
          </div>
          <div className="mb-2">
            If possible, use the NWC API directly! this API is subject to change
            or be removed entirely.
          </div>
          <Button size={"sm"} onClick={() => setShowToken(true)}>
            Show Token
          </Button>
        </AlertDescription>
      </Alert>
      {show && (
        <>
          <Textarea
            className="p-4 max-w-sm font-emoji break-all"
            value={getAuthToken() || ""}
            autoFocus
            onFocus={(e) => e.target.select()}
            rows={3}
          ></Textarea>
          <p className="font-bold text-xs">
            By default, the token expires in 30 days. Set the JWT_EXPIRY_DAYS
            enviroment variable if you need a longer expiry.
          </p>
          <p className="font-bold text-xs">
            This token can do anything with your hub! keep it safe! In case it
            somehow is leaked, change your JWT_SECRET environment variable.
          </p>
        </>
      )}
    </>
  );
}
