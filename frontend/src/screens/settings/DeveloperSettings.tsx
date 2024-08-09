import { Copy, Rocket, TriangleAlert } from "lucide-react";
import React from "react";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { getAuthToken } from "src/lib/auth";
import { copyToClipboard } from "src/lib/clipboard";

export default function DeveloperSettings() {
  const [show, setShowToken] = React.useState(false);
  const authToken = getAuthToken() || "";

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
            You can use your auth token to access the Alby Hub internal API.
            However, whenever possible, we recommend using the NWC API directly
            for more stability. Please note that the internal API may change or
            be removed entirely in the future.
          </div>
          {!show && (
            <Button size={"sm"} onClick={() => setShowToken(true)}>
              Show Token
            </Button>
          )}
          {show && (
            <>
              <div className="flex flex-row items-center gap-2 mb-2">
                <Input
                  type="password"
                  value={authToken}
                  className="flex-1"
                  readOnly
                />
                <Button
                  type="button"
                  variant="secondary"
                  size="icon"
                  onClick={() => {
                    copyToClipboard(authToken);
                  }}
                >
                  <Copy className="w-4 h-4" />
                </Button>
              </div>
              <p className="text-xs">
                By default, your token will expire in{" "}
                <span className="font-bold">30 days</span>. If you need a longer
                expiry period, you can adjust this by setting the
                JWT_EXPIRY_DAYS environment variable. This token grants full
                access to your hub. Please keep it secure. If you suspect that
                the token has been compromised, immediately change your
                JWT_SECRET environment variable to protect your data.
              </p>
            </>
          )}
        </AlertDescription>
      </Alert>
    </>
  );
}
